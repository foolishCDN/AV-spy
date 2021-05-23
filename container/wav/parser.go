package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/foolishCDN/AV-spy/encoding/riff"
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
	case bytes.Equal(chunk.ID, ChunkIDData):
		return p.handler.OnPCM(chunk.Data)
	case bytes.Equal(chunk.ID, ChunkIDWaveList):
		return p.parseWaveListChunk(chunk)
	case bytes.Equal(chunk.ID, ChunkIDSample):
		return p.parseSampleChunk(chunk)
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

func (p *Parser) parseSampleChunk(chunk *riff.Chunk) error {
	data := chunk.Data
	if len(data) < 36 {
		return errors.New("invalid sample chunk data")
	}
	sample := &MIDISample{
		Manufacturer:      binary.LittleEndian.Uint16(data[0:4]),
		Product:           binary.LittleEndian.Uint16(data[4:8]),
		SamplePeriod:      binary.LittleEndian.Uint16(data[8:12]),
		MIDIUnityNote:     binary.LittleEndian.Uint16(data[12:16]),
		MIDIPitchFraction: binary.LittleEndian.Uint16(data[16:20]),
		SMPTEFormat:       binary.LittleEndian.Uint16(data[20:24]),
		SMPTEOffset:       binary.LittleEndian.Uint16(data[24:28]),
		NumOfSampleLoops:  binary.LittleEndian.Uint16(data[28:32]),
		SampleData:        binary.LittleEndian.Uint16(data[32:36]),
	}
	data = data[36:]
	for i := uint16(0); i < sample.NumOfSampleLoops && len(data) >= 24; i++ {
		loop := MIDISampleLoop{
			ID:                   binary.LittleEndian.Uint16(data[0:4]),
			Type:                 binary.LittleEndian.Uint16(data[4:8]),
			Start:                binary.LittleEndian.Uint16(data[8:12]),
			End:                  binary.LittleEndian.Uint16(data[12:16]),
			Fraction:             binary.LittleEndian.Uint16(data[16:20]),
			NumOfTimesToPlayLoop: binary.LittleEndian.Uint16(data[20:24]),
		}
		data = data[24:]
		sample.SampleLoops = append(sample.SampleLoops, loop)
	}
	sample.SamplerSpecificData = data
	return p.handler.OnMIDISample(sample)
}

func (p *Parser) parseWaveListChunk(chunk *riff.Chunk) error {
	size := uint32(0)
	for i := uint32(0); len(chunk.Data[i:]) >= 8 && i < chunk.Size; i += size + 8 {
		id := chunk.Data[0:4]
		size = binary.LittleEndian.Uint32(chunk.Data[4:8])
		if len(chunk.Data[i+8:]) < int(size) {
			return errors.New("invalid wavl chunk data")
		}
		data := chunk.Data[i+8 : i+8+size]
		subChunk := &riff.Chunk{
			ID:   id,
			Size: size,
			Data: data,
		}
		if err := p.onChunk(subChunk); err != nil {
			return err
		}
	}
	if len(chunk.Data) > 0 {
		return errors.New("parse wavl chunk data error")
	}
	return nil
}
