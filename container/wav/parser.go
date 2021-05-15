package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/sikasjc/AV-spy/encoding/riff"
)

func NewParser(handler Handler) *Parser {
	return &Parser{
		handler: handler,
	}
}

type Parser struct {
	handler Handler

	p    *riff.Parser
	data []byte
}

func (p *Parser) Input(data []byte) error {
	if p.p == nil {
		p.p = new(riff.Parser)
		p.p.OnChunk = p.onChunk
		p.p.OnRIFFChunkHeader = p.onRIFFChunkHeader
	}
	return p.p.Input(data)
}

func (p *Parser) onChunk(chunk *riff.Chunk) error {
	switch {
	case bytes.Equal(chunk.ID, ChunkIDFMT):
		format, err := p.readFormat(chunk.Data)
		if err != nil {
			return err
		}
		return p.handler.OnFormat(format)
	case bytes.Equal(chunk.ID, ChunkIDDATA):
		return p.handler.OnPCM(chunk.Data)
	//case ChunkIDSample:
	//	// TODO
	default:
		if err := p.handler.OnUnknownChunk(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) onRIFFChunkHeader(header *riff.RIFFChunkHeader) error {
	if !bytes.Equal(header.FormatType, RIFFTypeIDWAVE) {
		return fmt.Errorf("parser: unsupport RIFF format type: %x", header.FormatType)
	}
	return nil
}

func (p *Parser) readFormat(data []byte) (*Format, error) {
	if len(data) < 16 {
		return nil, errors.New("parser: invalid format data")
	}
	format := &Format{
		Format:        binary.LittleEndian.Uint16(data[:2]),
		NumOfChannels: binary.LittleEndian.Uint16(data[2:4]),
		SampleRate:    binary.LittleEndian.Uint32(data[4:8]),
		ByteRate:      binary.LittleEndian.Uint32(data[8:12]),
		BlocKAlign:    binary.LittleEndian.Uint16(data[12:14]),
		BitPerSample:  binary.LittleEndian.Uint16(data[14:16]),
	}
	if len(data) > 18 {
		format.ExtraFormatSize = binary.LittleEndian.Uint16(data[16:18])
		if uint16(len(data[18:])) < format.ExtraFormatSize {
			return nil, errors.New("parser: invalid extra format data")
		}
		format.ExtraFormat = data[18:]
	}
	return format, nil
}
