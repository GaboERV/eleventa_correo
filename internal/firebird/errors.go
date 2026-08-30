package firebird

import (
	"strings"
	"unsafe"
)

type StatusError struct {
	Message string
}

func (e *StatusError) Error() string {
	return e.Message
}

func (fb *FBClient) InterpretError(statusVector *[20]uintptr) error {
	var msgs []string
	var buf [1024]byte
	var pVector uintptr = uintptr(unsafe.Pointer(&statusVector[0]))

	for {
		res, _, _ := fb.Interprete.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&pVector)),
		)
		if res == 0 {
			break
		}

		var end int
		for end = 0; end < 1024; end++ {
			if buf[end] == 0 {
				break
			}
		}

		if end > 0 {
			msgs = append(msgs, string(buf[:end]))
		}
	}

	return &StatusError{
		Message: strings.Join(msgs, "\n"),
	}
}
