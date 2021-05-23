package flv

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/foolishCDN/AV-spy/utils"
)

const (
	parseStateHeader = iota
	parseStateTagHeader
	parseStateTagAndPreviousTagSize
)

func NewParser(processPacket func(tag TagI) error) *Parser {
	return &Parser{
		processPacket: processPacket,
		demuxer:       new(Demuxer),
		headerBuf:     bytes.NewBuffer(make([]byte, 0, 9+4)),
		tagHeaderBuf:  bytes.NewBuffer(make([]byte, 0, 11)),
		tagBodyBuf:    bytes.NewBuffer(make([]byte, 0)),
	}
}

type Parser struct {
	Header *Header

	tagSize       int
	state         int
	processPacket func(tag TagI) error
	demuxer       *Demuxer

	headerBuf    *bytes.Buffer
	tagHeaderBuf *bytes.Buffer
	tagBodyBuf   *bytes.Buffer
}

func (p *Parser) Input(data []byte) (err error) {
	var n int

	total := len(data)
	for i := 0; i < total; i += n {
		switch p.state {
		case parseStateHeader:
			n = utils.Append(p.headerBuf, data[i:], 9+4)
			if p.headerBuf.Len() == 9+4 {
				headerBuf := p.headerBuf.Bytes()
				p.Header, err = p.demuxer.parseHeader(headerBuf[:9])
				if err != nil {
					return err
				}
				// previousTagSize0
				if binary.BigEndian.Uint32(headerBuf[9:9+4]) != 0 {
					return fmt.Errorf("flv demuxer previousTagSize0 should be 0, but there is: %v", binary.BigEndian.Uint32(headerBuf[9:9+4]))
				}
				p.state = parseStateTagHeader
			}
		case parseStateTagHeader:
			n = utils.Append(p.tagHeaderBuf, data[i:], 11)
			if p.tagHeaderBuf.Len() == 11 {
				tagHeaderBuf := p.tagHeaderBuf.Bytes()
				p.tagSize = int(utils.BigEndianUint24(tagHeaderBuf[1:4]))
				p.state = parseStateTagAndPreviousTagSize
			}
		case parseStateTagAndPreviousTagSize:
			n = utils.Append(p.tagBodyBuf, data[i:], p.tagSize+4)
			if p.tagBodyBuf.Len() == p.tagSize+4 {
				tagHeaderBuf := p.tagHeaderBuf.Bytes()
				tagBodyBuf := p.tagBodyBuf.Bytes()
				tag, err := p.demuxer.parseTag(uint32(p.tagSize), tagHeaderBuf, tagBodyBuf)
				if err != nil {
					return err
				}
				if err := p.processPacket(tag); err != nil {
					return err
				}
				p.state = parseStateTagHeader
				// reset
				p.tagHeaderBuf.Reset()
				p.tagBodyBuf.Reset()
			}
		default:
			return fmt.Errorf("flv parser invalid state: %d", p.state)
		}
	}
	return nil
}
