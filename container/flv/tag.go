package flv

import (
	"fmt"

	"github.com/foolishCDN/AV-spy/codec"
	"github.com/foolishCDN/AV-spy/codec/avc"
	"github.com/foolishCDN/AV-spy/codec/hevc"
	"github.com/foolishCDN/AV-spy/formatter"
	"github.com/foolishCDN/AV-spy/utils"
)

// Header FLV file header
type Header struct {
	Version    byte
	HasAudio   bool
	HasVideo   bool
	DataOffset uint32 // offset in bytes from start of file to start of body (that is, size of header)
}

// TagI ...
type TagI interface {
	Type() TagType
	Len() int
	Data() []byte
	Timestamp() uint32
}

// AudioTag ...
type AudioTag struct {
	SoundFormat  SoundFormat
	SampleRate   SoundRate // 0-5.5kHz, 1-11kHz, 2-22kHz, 3-44kHz
	BitPerSample SoundSize // 0-8bit, 1-16bit
	Channels     SoundType // 0-mono, 1-stereo

	PTS        uint32
	StreamID   uint32
	PacketType byte // 0-AAC sequence header, 1-AAC raw if SoundFormat=10
	Bytes      []byte
}

func (tag *AudioTag) Type() TagType {
	return TagAudio
}

func (tag *AudioTag) Len() int {
	size := len(tag.Bytes) + 1
	if tag.SoundFormat == AAC {
		size++
	}
	return size
}

func (tag *AudioTag) Data() []byte {
	return tag.Bytes
}

func (tag *AudioTag) Timestamp() uint32 {
	return tag.PTS
}

func (tag *AudioTag) ToVars() map[formatter.ElementName]interface{} {
	streamType := "AUDIO"
	if tag.SoundFormat == AAC && tag.PacketType == SequenceHeader {
		streamType = "AAC"
	}
	return map[formatter.ElementName]interface{}{
		formatter.ElementStreamType:        streamType,
		formatter.ElementStreamID:          tag.StreamID,
		formatter.ElementPTS:               tag.PTS,
		formatter.ElementDTS:               tag.PTS,
		formatter.ElementSize:              len(tag.Data()),
		formatter.ElementAudioSoundFormant: tag.SoundFormat.String(),
		formatter.ElementAudioChannels:     tag.Channels.String(),
		formatter.ElementAudioSampleRate:   tag.SampleRate.String(),
		formatter.ElementAudioSoundSize:    tag.BitPerSample.String(),
	}
}

// VideoTag ...
type VideoTag struct {
	FrameType FrameType
	CodecID   CodecID

	DTS        uint32
	PTS        uint32
	StreamID   uint32
	PacketType byte // 0-AVC sequence header, 1-AVC NALU, 2-AVC end of sequence if CodecID=7
	Bytes      []byte

	NALUs    [][]byte
	NALUType codec.NALUType
}

func (tag *VideoTag) Type() TagType {
	return TagVideo
}

func (tag *VideoTag) Len() int {
	size := len(tag.Bytes) + 1
	if tag.CodecID == H264 || tag.CodecID == H265 {
		size += 4
	}
	return size
}

func (tag *VideoTag) Data() []byte {
	return tag.Bytes
}

func (tag *VideoTag) Timestamp() uint32 {
	return tag.DTS
}

func (tag *VideoTag) NALUTypes() ([]uint8, string) {
	switch tag.CodecID {
	case H264:
		var nalus [][]byte
		var t avc.NALUType
		if tag.PacketType == SequenceHeader {
			nalus = tag.NALUs
			t = avc.NALUTypeAVCC
		} else {
			if len(tag.NALUs) == 0 {
				nalus, t = avc.SplitNALUs(tag.Data())
				tag.NALUs = nalus
				tag.NALUType = t
			} else {
				nalus = tag.NALUs
				t = tag.NALUType.(avc.NALUType)
			}
		}
		if t == avc.NALUTypeInvalid {
			return nil, avc.NALUTypeInvalid.String()
		}
		naluTypes := make([]uint8, 0, len(nalus))
		for _, nalu := range nalus {
			reader := utils.NewBitReader(nalu)
			header := avc.ParseNALUHeader(reader)
			naluTypes = append(naluTypes, header.NalUnitType)
		}
		return naluTypes, t.String()
	case H265:
		var nalus [][]byte
		var t hevc.NALUType
		if tag.PacketType == SequenceHeader {
			nalus = tag.NALUs
			t = hevc.NALUTypeHVCC
		} else {
			if len(tag.NALUs) == 0 {
				nalus, t = hevc.SplitNALUS(tag.Data())
				tag.NALUs = nalus
				tag.NALUType = t
			} else {
				nalus = tag.NALUs
				t = tag.NALUType.(hevc.NALUType)
			}
		}
		if t == hevc.NALUTypeInvalid {
			return nil, hevc.NALUTypeInvalid.String()
		}
		naluTypes := make([]uint8, 0, len(nalus))
		for _, nalu := range nalus {
			reader := utils.NewBitReader(nalu)
			header := hevc.ParseNALUHeader(reader)
			naluTypes = append(naluTypes, header.NALUnitType)
		}
		return naluTypes, t.String()
	}
	return nil, "unsupported"
}

func (tag *VideoTag) SEI() (int, int, []byte) {
	switch tag.CodecID {
	case H264:
		naluTypes, _ := tag.NALUTypes()
		for i, naluType := range naluTypes {
			if naluType != avc.NalSEI {
				continue
			}
			nalu := tag.NALUs[i]
			reader := utils.NewBitReader(nalu)
			avc.ParseNALUHeader(reader)
			payloadType := reader.Read8BitsUntilNot0xFF()
			payloadSize := reader.Read8BitsUntilNot0xFF()
			return payloadType, payloadSize, reader.LastData()[:payloadSize]
		}
	case H265:
		naluTypes, _ := tag.NALUTypes()
		for i, naluType := range naluTypes {
			if naluType != hevc.NalSEIPrefix && naluType != hevc.NalSEISuffix {
				continue
			}
			nalu := tag.NALUs[i]
			reader := utils.NewBitReader(nalu)
			hevc.ParseNALUHeader(reader)
			payloadType := reader.Read8BitsUntilNot0xFF()
			payloadSize := reader.Read8BitsUntilNot0xFF()
			return payloadType, payloadSize, reader.LastData()[:payloadSize]
		}
	}
	return 0, 0, nil
}

func (tag *VideoTag) ToVars() map[formatter.ElementName]interface{} {
	streamType := "VIDEO"
	if tag.PacketType == SequenceHeader {
		streamType = "AVC"
		if tag.CodecID == H265 {
			streamType = "HEVC"
		}
	}
	naluTypes, t := tag.NALUTypes()
	return map[formatter.ElementName]interface{}{
		formatter.ElementStreamType:     streamType,
		formatter.ElementStreamID:       tag.StreamID,
		formatter.ElementPTS:            tag.PTS,
		formatter.ElementDTS:            tag.DTS,
		formatter.ElementSize:           len(tag.Data()),
		formatter.ElementNALUTypes:      fmt.Sprintf("%s %v", t, naluTypes),
		formatter.ElementVideoFrameType: tag.FrameType.String(),
		formatter.ElementVideoCodecID:   tag.CodecID.String(),
	}
}

// ScriptTag ...
type ScriptTag struct {
	PTS      uint32
	StreamID uint32
	Bytes    []byte
}

func (tag *ScriptTag) Type() TagType {
	return TagScript
}

func (tag *ScriptTag) Len() int {
	return len(tag.Bytes)
}

func (tag *ScriptTag) Data() []byte {
	return tag.Bytes
}

func (tag *ScriptTag) Timestamp() uint32 {
	return tag.PTS
}

func (tag *ScriptTag) ToVars() map[formatter.ElementName]interface{} {
	return map[formatter.ElementName]interface{}{
		formatter.ElementStreamType: "SCRIPT",
		formatter.ElementStreamID:   tag.StreamID,
		formatter.ElementPTS:        tag.PTS,
		formatter.ElementDTS:        tag.PTS,
		formatter.ElementSize:       len(tag.Data()),
	}
}
