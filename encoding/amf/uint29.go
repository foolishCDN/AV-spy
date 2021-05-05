package amf

import (
	"errors"
	"io"
)

func DecodeUint29(r io.Reader) (n uint32, err error) {
	buf := make([]byte, 1)

	for i := 0; i < 3; i++ {
		if _, err := r.Read(buf); err != nil {
			return 0, err
		}
		n = (n << 7) + uint32(buf[0]&0x7F)
		if buf[0]&0x80 == 0 {
			return n, nil
		}
	}
	if _, err := r.Read(buf); err != nil {
		return 0, err
	}
	n = n<<8 + uint32(buf[0])
	return n, nil
}

func EncodeUint29(w io.Writer, n uint32) error {
	var b []byte
	switch {
	case n <= 0x0000007F:
		b = make([]byte, 1)
		b[0] = byte(n)
	case n <= 0x00003FFF:
		b = make([]byte, 2)
		b[0] = byte(n>>7 | 0x80)
		b[1] = byte(n & 0x7F)
	case n <= 0x001FFFFF:
		b = make([]byte, 3)
		b[0] = byte(n>>14 | 0x80)
		b[1] = byte(n>>7&0x7F | 0x80)
		b[2] = byte(n & 0x7F)
	case n <= 0x1FFFFFFF:
		b = make([]byte, 4)
		b[0] = byte(n>>22 | 0x80)
		b[1] = byte(n>>15&0x7F | 0x80)
		b[2] = byte(n>>8&0x7F | 0x80)
		b[3] = byte(n)
	default:
		return errors.New("out of range")
	}
	_, err := w.Write(b)
	return err
}
