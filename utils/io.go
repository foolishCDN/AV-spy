package utils

import (
	"bytes"
	"io"
)

// WriteFull is similar to io.ReadFull, here it is write.
func WriteFull(w io.Writer, data []byte) (err error) {
	var n int
	for i := 0; i < len(data); i += n {
		if n, err = w.Write(data[i:]); err != nil {
			return err
		}
	}
	return nil
}

// Append appends data to buf to len(buf) == n, returns the num of appended data and buf
func Append(buf *bytes.Buffer, data []byte, n int) int {
	now := buf.Len()
	if now+len(data) <= n {
		buf.Write(data)
		return len(data)
	}
	buf.Write(data[:n-now])
	return n - now
}
