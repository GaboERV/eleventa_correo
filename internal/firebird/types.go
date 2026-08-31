package firebird

import (
	"time"
)

const (
	SQL_TEXT      = 452
	SQL_VARYING   = 448
	SQL_SHORT     = 500
	SQL_LONG      = 496
	SQL_FLOAT     = 482
	SQL_DOUBLE    = 480
	SQL_TIMESTAMP = 510
	SQL_BLOB      = 520
	SQL_INT64     = 580

	DSQL_close = 1
	DSQL_drop  = 2
	MAX_COLS   = 100

	SQLDA_VERSION1 = 1
)

type ISC_QUAD struct {
	GsecQuadHigh int32
	GsecQuadLow  uint32
}

type XSQLVAR struct {
	SqlType    int16
	SqlScale   int16
	SqlSubtype int16
	SqlLen     int16
	SqlData    uintptr
	SqlInd     uintptr
	SqlNameLen int16
	SqlName    [32]byte
	RelNameLen int16
	RelName    [32]byte
	OwnNameLen int16
	OwnName    [32]byte
	AliasLen   int16
	AliasName  [32]byte
}

type XSQLDA struct {
	Version int16
	DaName  [8]byte
	DaBc    int32
	SqlNum  int16
	SqlD    int16
	SqlVar  [MAX_COLS]XSQLVAR
}

type ISC_TIMESTAMP struct {
	Date int32
	Time uint32
}

func TimeToISCTimestamp(t time.Time) ISC_TIMESTAMP {
	// Date calculation (MJD)
	y, m, d := t.Date()
	year := int(y)
	month := int(m)
	day := int(d)

	if month <= 2 {
		year--
		month += 12
	}
	mjd := 365*year + year/4 - year/100 + year/400 + (153*(month-3)+2)/5 + day - 678882

	// Time calculation (deci-milliseconds since midnight)
	h, min, s := t.Clock()
	ns := t.Nanosecond()
	deciMs := uint32(h)*36000000 + uint32(min)*600000 + uint32(s)*10000 + uint32(ns/100000)

	return ISC_TIMESTAMP{
		Date: int32(mjd),
		Time: deciMs,
	}
}

func ISCTimestampToTime(ts ISC_TIMESTAMP, loc *time.Location) time.Time {
	// MJD to Gregorian (0 = Nov 17, 1858)
	mjd := int(ts.Date)
	jd := mjd + 2400001
	a := jd + 32044
	b := (4*a + 3) / 146097
	c = a - (146097*b)/4
	d := (4*c + 3) / 1461
	e := c - (1461*d)/4
	m := (5*e + 2) / 153
	day := e - (153*m+2)/5 + 1
	month := m + 3 - 12*(m/10)
	year := b*100 + d - 4800 + m/10

	// Deci-milliseconds to time
	deciMs := ts.Time
	h := int(deciMs / 36000000)
	deciMs %= 36000000
	min := int(deciMs / 600000)
	deciMs %= 600000
	s := int(deciMs / 10000)
	deciMs %= 10000
	ns := int(deciMs) * 100000

	if loc == nil {
		loc = time.UTC
	}

	return time.Date(year, time.Month(month), day, h, min, s, ns, loc)
}
