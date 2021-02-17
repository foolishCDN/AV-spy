package utils

import (
	"bytes"
)

// BigEndianPutUint24 is similar to binary.BigEndian.PutUint32, here it is uint24.
func BigEndianPutUint24(buf []byte, v uint32) {
	_ = buf[2]
	buf[0] = byte(v >> 16)
	buf[1] = byte(v >> 8)
	buf[2] = byte(v)
}

// BigEndianUint24 is similar to binary.BigEndian.Uint32, here it is uint24.
func BigEndianUint24(b []byte) uint32 {
	_ = b[2]
	return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
}

func BigEndianPutUint32ToBuffer(b *bytes.Buffer, v uint32) {
	b.WriteByte(byte(v >> 24))
	b.WriteByte(byte(v >> 16))
	b.WriteByte(byte(v >> 8))
	b.WriteByte(byte(v))
}

func BigEndianPutUint24ToBuffer(b *bytes.Buffer, v uint32) {
	b.WriteByte(byte(v >> 16))
	b.WriteByte(byte(v >> 8))
	b.WriteByte(byte(v))
}

func BigEndianPutUint16ToBuffer(b *bytes.Buffer, v uint16) {
	b.WriteByte(byte(v >> 8))
	b.WriteByte(byte(v))
}

func LittleEndianPutUint32ToBuffer(b *bytes.Buffer, v uint32) {
	b.WriteByte(byte(v))
	b.WriteByte(byte(v >> 8))
	b.WriteByte(byte(v >> 16))
	b.WriteByte(byte(v >> 24))
}
