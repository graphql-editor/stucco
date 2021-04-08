package utils

import (
	"bytes"
	"io"
)

// CopyToReader returns new bytes.Reader with a copy of data read by r
func CopyToReader(r io.Reader) (br *bytes.Reader, err error) {
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err == nil {
		br = bytes.NewReader(buf.Bytes())
	}
	return
}

// ReaderAtToReader returns a reader backed with ReaderAt
func ReaderAtToReader(r io.ReaderAt, size int64) io.Reader {
	return io.NewSectionReader(r, 0, size)
}
