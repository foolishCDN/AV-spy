package flv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/foolishCDN/AV-spy/utils"
)

type Demuxer struct {
	readTagHeaderBuf [11]byte
}

// ReadHeader read flv file header
//
// FLV File Header (9 byte)
//     signature (3 byte) 'F', 'L', 'V'
//     version (1 byte) 1
//     flags (1 byte)
//     offset (4 byte) the length of FLV header
func (demuxer *Demuxer) ReadHeader(r io.Reader) (*Header, error) {
	header := make([]byte, 9)
	if _, err := io.ReadFull(r, header[0:9]); err != nil {
		return nil, err
	}
	h, err := demuxer.parseHeader(header)
	if err != nil {
		return nil, err
	}

	// previousTagSize0
	temp := header[:4]
	if _, err := io.ReadFull(r, temp); err != nil {
		return nil, err
	}
	if binary.BigEndian.Uint32(temp) != 0 {
		return nil, fmt.Errorf("flv demuxer previousTagSize0 should be 0, but there is: %v", binary.BigEndian.Uint32(temp))
	}
	return h, nil
}

func (demuxer *Demuxer) parseHeader(header []byte) (*Header, error) {
	if bytes.Equal(header[:2], Signature) {
		return nil, fmt.Errorf("flv signature doesn't match, %q!=FLV", header[:2])
	}
	h := &Header{
		Version:    header[3],
		HasAudio:   (header[4] & 0x04) != 0,
		HasVideo:   (header[4] & 0x01) != 0,
		DataOffset: binary.BigEndian.Uint32(header[5:9]),
	}
	return h, nil
}

// ReadTag read flv tag
//
// FLV Tag (11 + len(data))
//     type (1 byte)
//     size (3 byte) start from streamID
//     timestamp (3 byte)
//     timestampExtended (1 byte)
//     streamID (3 byte) always 0
//     data
func (demuxer *Demuxer) ReadTag(r io.Reader) (TagI, error) {
	tagHeader := demuxer.readTagHeaderBuf[:]
	if _, err := io.ReadFull(r, tagHeader[:11]); err != nil {
		return nil, err
	}
	size := utils.BigEndianUint24(tagHeader[1:4])
	data := make([]byte, size+4) // has previousTagSizeN
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	return demuxer.parseTag(size, tagHeader, data)
}
func (demuxer *Demuxer) parseTag(size uint32, tagHeader []byte, data []byte) (TagI, error) {
	if size+11 != binary.BigEndian.Uint32(data[size:]) { // Verified by previousTagSizeN
		return nil, fmt.Errorf("flv demuxer read tag size %d + 11 != %d", size, binary.BigEndian.Uint32(data[size:]))
	}

	tag, err := demuxer.demux(
		tagHeader[0]&0xa0,                      // filter
		TagType(tagHeader[0]&0x1f),             // tag type
		utils.BigEndianUint24(tagHeader[8:11]), // streamID
		(uint32(tagHeader[4])<<16)|uint32(tagHeader[5])<<8|uint32(tagHeader[6])|uint32(tagHeader[7])<<24, // timestamp
		data[:size], // data
	)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (demuxer *Demuxer) demux(filter byte, tagType TagType, streamID, timestamp uint32, data []byte) (t TagI, err error) {
	switch tagType {
	case TagAudio:
		t = demuxer.audioTag(data, streamID, timestamp)
	case TagVideo:
		t = demuxer.videoTag(data, streamID, timestamp)
	case TagScript:
		t = demuxer.scriptTag(data, streamID, timestamp)
	default:
		return nil, fmt.Errorf("flv demuxer unknown tag type %d", tagType)
	}
	return t, nil
}

func (demuxer *Demuxer) audioTag(data []byte, streamID, timestamp uint32) *AudioTag {
	a := &AudioTag{}
	a.PTS = timestamp
	a.StreamID = streamID
	a.SoundFormat = SoundFormat((data[0] >> 4) & 0x0f)
	a.SampleRate = SoundRate((data[0] >> 2) & 0x03)
	a.BitPerSample = SoundSize((data[0] >> 1) & 0x01)
	a.Channels = SoundType(data[0] & 0x01)

	if a.SoundFormat == AAC {
		a.PacketType = data[1]
		a.Bytes = data[2:]
	} else {
		a.Bytes = data[1:]
	}
	return a
}

func (demuxer *Demuxer) videoTag(data []byte, streamID, timestamp uint32) *VideoTag {
	v := &VideoTag{}
	v.DTS = timestamp
	v.StreamID = streamID
	v.FrameType = Format((data[0] >> 4) & 0xf)
	v.CodecID = CodecID(data[0] & 0xf)

	switch v.CodecID {
	case H264, H265:
		v.PacketType = data[1]
		compositionTime := utils.BigEndianUint24(data[2:5])
		compositionTime = (compositionTime + 0xFF800000) ^ 0xFF800000
		v.PTS = uint32(int64(v.DTS) + int64(int32(compositionTime)))
		v.Bytes = data[5:]

	default:
		v.Bytes = data[1:]
	}
	return v
}

func (demuxer *Demuxer) scriptTag(data []byte, streamID, timestamp uint32) *ScriptTag {
	s := &ScriptTag{}
	s.PTS = timestamp
	s.StreamID = streamID
	s.Bytes = data
	return s
}
