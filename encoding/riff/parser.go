package riff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/foolishCDN/AV-spy/utils"
)

type Parser struct {
	OnChunk           func(chunk *Chunk) error
	OnRIFFChunkHeader func(header *RIFFChunkHeader) error

	buf     bytes.Buffer
	state   int
	id      [4]byte
	total   uint32
	size    uint32
	padding uint32
}

func (p *Parser) Input(data []byte) error {
	var n int
	total := len(data)
	for i := 0; i < total; i += n {
		switch p.state {
		case StateReadRIFFChunkHeader:
			n = utils.Append(&p.buf, data[i:], 12)
			if p.buf.Len() == 12 {
				buf := p.buf.Bytes()
				if !bytes.Equal(buf[0:4], ChunkIDRIFF) {
					return fmt.Errorf("chunk id should be 'RIFF', got %q", p.buf.Bytes())
				}
				p.total = binary.LittleEndian.Uint32(buf[4:8])
				formatType := buf[8:12]
				if err := p.OnRIFFChunkHeader(&RIFFChunkHeader{Size: p.total, FormatType: formatType}); err != nil {
					return err
				}
				p.total -= 4 // riff chunk size = formatType + chunks, so chunks = riff chunk size - 4
				p.state = StateReadChunkHeader
				p.buf.Reset()
			}
		case StateReadChunkHeader:
			n = utils.Append(&p.buf, data[i:], 8)
			if p.buf.Len() == 8 {
				buf := p.buf.Bytes()
				copy(p.id[:], buf[0:4])
				p.size = binary.LittleEndian.Uint32(buf[4:8])
				if p.size%2 == 1 {
					p.padding = 1
				}
				p.state = StateReadChunkData
				p.buf.Reset()
			}
		case StateReadChunkData:
			n = utils.Append(&p.buf, data[i:], int(p.size+p.padding))
			if p.buf.Len() == int(p.size+p.padding) {
				chunk := &Chunk{
					ID:   make([]byte, 4),
					Size: p.size,
					Data: p.buf.Bytes(),
				}
				copy(chunk.ID, p.id[:])
				if err := p.OnChunk(chunk); err != nil {
					return err
				}
				p.total = p.total - (p.size + p.padding)
				if p.total <= 0 {
					p.state = StateReadDone
				}

				// reset
				p.state = StateReadChunkHeader
				p.buf.Reset()
				p.padding = 0
			}
		case StateReadDone:
			return io.EOF
		}
	}
	return nil
}
