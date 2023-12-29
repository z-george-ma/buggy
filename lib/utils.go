package lib

import (
	"context"
	"unsafe"
)

// IsDone returns true if ctx is complete
func IsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func BytesToString(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
