package firebird

import (
	"fmt"
	"unsafe"
)

type Connection struct {
	dbHandle uintptr
	fb       *FBClient
}

type Transaction struct {
	trHandle  uintptr
	fb        *FBClient
	committed bool
}

func (fb *FBClient) Connect(dbPath, user, pass string) (*Connection, error) {
	var statusVector [20]uintptr
	var dbHandle uintptr = 0

	dbNameBytes := append([]byte(dbPath), 0)
	dpb := []byte{1, 28, byte(len(user))}
	dpb = append(dpb, []byte(user)...)
	dpb = append(dpb, 29, byte(len(pass)))
	dpb = append(dpb, []byte(pass)...)

	ret, _, _ := fb.AttachDatabase.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(len(dbPath)),
		uintptr(unsafe.Pointer(&dbNameBytes[0])),
		uintptr(unsafe.Pointer(&dbHandle)),
		uintptr(len(dpb)),
		uintptr(unsafe.Pointer(&dpb[0])),
	)

	if ret != 0 {
		return nil, fmt.Errorf("failed to attach database: %w", fb.InterpretError(&statusVector))
	}

	return &Connection{
		dbHandle: dbHandle,
		fb:       fb,
	}, nil
}

func (c *Connection) Close() error {
	if c.dbHandle == 0 {
		return nil
	}
	var statusVector [20]uintptr
	ret, _, _ := c.fb.DetachDatabase.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&c.dbHandle)),
	)
	if ret != 0 {
		return c.fb.InterpretError(&statusVector)
	}
	c.dbHandle = 0
	return nil
}

func (c *Connection) BeginTransaction() (*Transaction, error) {
	var statusVector [20]uintptr
	var trHandle uintptr = 0

	ret, _, _ := c.fb.StartTransaction.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&trHandle)),
		1,
		uintptr(unsafe.Pointer(&c.dbHandle)),
		0,
		0,
	)

	if ret != 0 {
		return nil, fmt.Errorf("failed to start transaction: %w", c.fb.InterpretError(&statusVector))
	}

	return &Transaction{
		trHandle: trHandle,
		fb:       c.fb,
	}, nil
}

func (tx *Transaction) Commit() error {
	if tx.trHandle == 0 || tx.committed {
		return nil
	}
	var statusVector [20]uintptr
	ret, _, _ := tx.fb.CommitTransaction.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&tx.trHandle)),
	)
	if ret != 0 {
		return tx.fb.InterpretError(&statusVector)
	}
	tx.committed = true
	tx.trHandle = 0
	return nil
}

func (tx *Transaction) Rollback() error {
	if tx.trHandle == 0 || tx.committed {
		return nil
	}
	var statusVector [20]uintptr
	ret, _, _ := tx.fb.RollbackTransaction.Call(
		uintptr(unsafe.Pointer(&statusVector[0])),
		uintptr(unsafe.Pointer(&tx.trHandle)),
	)
	if ret != 0 {
		return tx.fb.InterpretError(&statusVector)
	}
	tx.trHandle = 0
	return nil
}
