package mp3

import (
	"bytes"
	"io"
)

func NewDecoder(r io.Reader) (*Decoder, error) {
	decoder := &Decoder{}
	return decoder, nil
}

// Decoder is a simple MP3 decoder.
type Decoder struct {
	buf [3]byte
}

// ReadTags reads Tag ver.1(ID3V1) or Tag ver.2(ID3V2) from the MP3 file.
// TAG ver.1 takes always 128 Bytes at the very end of file (after the last audio frame).
// TAG ver.2 is placed at the very beginning of file (before all audio frames).
func (decoder *Decoder) ReadTags(r io.Reader) error {
	buf := make([]byte, 3)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	switch {
	case bytes.Equal(buf, TagVer1Identifier):
		b := make([]byte, 125)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
	case bytes.Equal(buf, TagVer2Identifier):
		b := make([]byte, 3)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
		b = make([]byte, 4)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
		tagSize := (uint32(b[0]) << 21) | (uint32(b[1]) << 14) | (uint32(b[2]) << 7) | (uint32(b[3]))
		b = make([]byte, tagSize)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
	default:
		copy(decoder.buf[:], buf)
	}
	return nil
}

func (decoder *Decoder) ReadFrame(r io.Reader) error {
	// TODO
	return nil
}
