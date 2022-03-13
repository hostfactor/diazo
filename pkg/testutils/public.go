package testutils

import (
	"bytes"
	"runtime"
)

func GetCurrentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	return filename
}

type ByteBuffer struct {
	bytes.Buffer
}

func (b *ByteBuffer) Close() error {
	return nil
}
