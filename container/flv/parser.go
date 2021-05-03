package flv

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/sikasjc/AV-spy/utils"
)

const (
	parseStateHeader = iota
	parseStateTagHeader
	parseStateTagAndPreviousTagSize
)

type HandlerI interface {
	OnPacket(tag TagI) error
}

func NewParser(hander HandlerI) *Parser {
	return &Parser{
		handler:      hander,
		demuxer:      new(Demuxer),
		headerBuf:    bytes.NewBuffer(make([]byte, 0, 9+4)),
		tagHeaderBuf: bytes.NewBuffer(make([]byte, 0, 11)),
		tagBodyBuf:   bytes.NewBuffer(make([]byte, 0)),
	}
}

type Parser struct {
	Header *Header

	state   int
	handler HandlerI
	demuxer *Demuxer

	headerBuf    *bytes.Buffer
	tagHeaderBuf *bytes.Buffer
	tagBodyBuf   *bytes.Buffer
}

func (p *Parser) Input(data []byte) (err error) {
	var n int
	var tagSize int

	total := len(data)
	for i := 0; i < total; i += n {
		switch p.state {
		case parseStateHeader:
			n = utils.Append(p.headerBuf, data, 9+4)
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
			n = utils.Append(p.tagHeaderBuf, data, 11)
			if p.tagHeaderBuf.Len() == 11 {
				tagHeaderBuf := p.tagHeaderBuf.Bytes()
				tagSize = int(binary.BigEndian.Uint32(tagHeaderBuf[1:4]))
				p.state = parseStateTagAndPreviousTagSize
			}
		case parseStateTagAndPreviousTagSize:
			n = utils.Append(p.tagBodyBuf, data, tagSize+11)
			if p.tagBodyBuf.Len() == tagSize+11 {
				tagHeaderBuf := p.tagHeaderBuf.Bytes()
				tagBodyBuf := p.tagBodyBuf.Bytes()
				tag, err := p.demuxer.parseTag(uint32(tagSize), tagHeaderBuf, tagBodyBuf)
				if err != nil {
					return err
				}
				if err := p.handler.OnPacket(tag); err != nil {
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
