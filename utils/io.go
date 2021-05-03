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

// BigEndianPutUint24 is similar to binary.BigEndian.PutUint32, here it is uint24.
func BigEndianPutUint24(buf []byte, v uint32) {
	_ = buf[2] // early bounds check to guarantee safety of writes below
	buf[0] = byte(v >> 16)
	buf[1] = byte(v >> 8)
	buf[2] = byte(v)
}

// BigEndianUint24 is similar to binary.BigEndian.Uint32, here it is uint24.
func BigEndianUint24(b []byte) uint32 {
	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
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
