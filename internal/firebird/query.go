package firebird

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"
	"unsafe"
)

type ResultSet struct {
	stmtHandle uintptr
	sqlda      *XSQLDA
	buffers    [][]byte
	nulls      []int16
	fb         *FBClient
	closed     bool
}

func (c *Connection) Query(tx *Transaction, sql string, params ...interface{}) (*ResultSet, error) {
	var statusVector [20]uintptr
	var stmtHandle uintptr = 0

	// 1. Allocate Statement
	retAlloc, _, _ := c.fb.DsqlAllocateStatement.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&c.dbHandle)),
		uintptr(unsafe.Pointer(&stmtHandle)),
	)
	if retAlloc != 0 {
		return nil, fmt.Errorf("allocate stmt failed: %w", c.fb.InterpretError(&statusVector))
	}

	// Helper to free on failure
	cleanup := func() {
		if stmtHandle != 0 {
			c.fb.DsqlFreeStatement.Call(uintptr(unsafe.Pointer(&statusVector[0])), uintptr(unsafe.Pointer(&stmtHandle)), DSQL_drop)
		}
	}

	// 2. Prepare Query
	queryBytes := append([]byte(sql), 0)
	retPrep, _, _ := c.fb.DsqlPrepare.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&tx.trHandle)),
		uintptr(unsafe.Pointer(&stmtHandle)),
		0,
		uintptr(unsafe.Pointer(&queryBytes[0])),
		3,
		0,
	)
	if retPrep != 0 {
		cleanup()
		return nil, fmt.Errorf("prepare failed: %w", c.fb.InterpretError(&statusVector))
	}

	// 3. Describe Output
	sqlda := &XSQLDA{
		Version: SQLDA_VERSION1,
		SqlNum:  MAX_COLS,
	}
	retDesc, _, _ := c.fb.DsqlDescribe.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&stmtHandle)),
		SQLDA_VERSION1,
		uintptr(unsafe.Pointer(sqlda)),
	)
	if retDesc != 0 {
		cleanup()
		return nil, fmt.Errorf("describe failed: %w", c.fb.InterpretError(&statusVector))
	}

	cols := int(sqlda.SqlD)
	if cols > MAX_COLS {
		cols = MAX_COLS // In a real app we'd reallocate XSQLDA, but MAX_COLS is enough
	}

	buffers := make([][]byte, cols)
	nulls := make([]int16, cols)

	for i := 0; i < cols; i++ {
		baseType := sqlda.SqlVar[i].SqlType &^ 1
		// Make it nullable
		sqlda.SqlVar[i].SqlType = baseType + 1

		var bufSize int16
		switch baseType {
		case SQL_SHORT:
			bufSize = 2
		case SQL_LONG, SQL_FLOAT:
			bufSize = 4
		case SQL_INT64, SQL_DOUBLE, SQL_TIMESTAMP, SQL_BLOB:
			bufSize = 8
		case SQL_TEXT:
			bufSize = sqlda.SqlVar[i].SqlLen
		case SQL_VARYING:
			bufSize = sqlda.SqlVar[i].SqlLen + 2
		default:
			bufSize = sqlda.SqlVar[i].SqlLen
			if bufSize < 256 {
				bufSize = 256
			}
		}

		sqlda.SqlVar[i].SqlLen = bufSize
		buffers[i] = make([]byte, bufSize)
		sqlda.SqlVar[i].SqlData = uintptr(unsafe.Pointer(&buffers[i][0]))
		sqlda.SqlVar[i].SqlInd = uintptr(unsafe.Pointer(&nulls[i]))
	}

	// 4. Bind Input Parameters if any
	var inSqldaPtr uintptr = 0
	var inputTimestamps []ISC_TIMESTAMP
	var inputNulls []int16
	if len(params) > 0 {
		inSqlda := &XSQLDA{
			Version: SQLDA_VERSION1,
			SqlNum:  int16(len(params)),
			SqlD:    int16(len(params)),
		}
		inputTimestamps = make([]ISC_TIMESTAMP, len(params))
		inputNulls = make([]int16, len(params))

		for i, param := range params {
			inSqlda.SqlVar[i].SqlType = SQL_TIMESTAMP + 1
			inSqlda.SqlVar[i].SqlLen = 8
			
			if t, ok := param.(time.Time); ok {
				inputTimestamps[i] = TimeToISCTimestamp(t)
				inputNulls[i] = 0
			} else {
				inputNulls[i] = -1 // Null
			}
			
			inSqlda.SqlVar[i].SqlData = uintptr(unsafe.Pointer(&inputTimestamps[i]))
			inSqlda.SqlVar[i].SqlInd = uintptr(unsafe.Pointer(&inputNulls[i]))
		}
		inSqldaPtr = uintptr(unsafe.Pointer(inSqlda))
	}

	// 5. Execute
	retExec, _, _ := c.fb.DsqlExecute.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&tx.trHandle)),
		uintptr(unsafe.Pointer(&stmtHandle)),
		SQLDA_VERSION1,
		inSqldaPtr,
	)
	if retExec != 0 {
		cleanup()
		return nil, fmt.Errorf("execute failed: %w", c.fb.InterpretError(&statusVector))
	}

	return &ResultSet{
		stmtHandle: stmtHandle,
		sqlda:      sqlda,
		buffers:    buffers,
		nulls:      nulls,
		fb:         c.fb,
	}, nil
}

func (rs *ResultSet) Next() bool {
	var statusVector [20]uintptr
	ret, _, _ := rs.fb.DsqlFetch.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&rs.stmtHandle)),
		SQLDA_VERSION1,
		uintptr(unsafe.Pointer(rs.sqlda)),
	)
	// 100 means EOF
	return ret == 0
}

func (rs *ResultSet) Close() {
	if rs.closed {
		return
	}
	var statusVector [20]uintptr
	rs.fb.DsqlFreeStatement.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&rs.stmtHandle)),
		DSQL_drop,
	)
	rs.closed = true
}

func (rs *ResultSet) isNull(col int) bool {
	return rs.nulls[col] == -1
}

func (rs *ResultSet) ScanInt(col int) (int, bool) {
	if rs.isNull(col) {
		return 0, true
	}
	val := *(*int32)(unsafe.Pointer(&rs.buffers[col][0]))
	return int(val), false
}

func (rs *ResultSet) ScanInt64(col int) (int64, bool) {
	if rs.isNull(col) {
		return 0, true
	}
	val := *(*int64)(unsafe.Pointer(&rs.buffers[col][0]))
	return val, false
}

func (rs *ResultSet) ScanString(col int) (string, bool) {
	if rs.isNull(col) {
		return "", true
	}
	val := strings.TrimSpace(string(bytes.Trim(rs.buffers[col], "\x00")))
	return val, false
}

func (rs *ResultSet) ScanTime(col int, loc *time.Location) (time.Time, bool) {
	if rs.isNull(col) {
		return time.Time{}, true
	}
	ts := *(*ISC_TIMESTAMP)(unsafe.Pointer(&rs.buffers[col][0]))
	return ISCTimestampToTime(ts, loc), false
}

func (rs *ResultSet) ScanBool(col int) (bool, bool) {
	if rs.isNull(col) {
		return false, true
	}
	val := *(*int16)(unsafe.Pointer(&rs.buffers[col][0]))
	return val != 0, false
}

// ScanCentavos reads a scaled numeric value (like ACUMULADO_VENTAS) and returns it in centavos.
func (rs *ResultSet) ScanCentavos(col int) (int64, bool) {
	if rs.isNull(col) {
		return 0, true
	}
	
	// Firebird NUMERIC(15,2) or similar can be stored as INT64
	val := *(*int64)(unsafe.Pointer(&rs.buffers[col][0]))
	scale := rs.sqlda.SqlVar[col].SqlScale // typically negative, e.g., -4 for NUMERIC(15,4)
	
	// Convert to float64 to apply Firebird scale
	fval := float64(val) * math.Pow10(int(scale))
	
	// Convert float to centavos (multiply by 100 and round to int)
	centavos := int64(math.Round(fval * 100))
	return centavos, false
}
