package flv

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sikasjc/AV-spy/utils"
)

type Muxer struct {
	writeTagHeaderBuf [11 + 4]byte
	muxerAudioTagBuf  [2]byte
	muxerVideoTagBuf  [5]byte
}

// WriteHeader sends FLV file header.
func (muxer *Muxer) WriteHeader(w io.Writer, hasAudio, hasVideo bool) error {
	header := make([]byte, 13) // 9+4 header + previousTagSize 0
	// signature
	header[0] = 'F'
	header[1] = 'L'
	header[2] = 'V'
	// version
	header[3] = 0x01
	// type flags (audio/video)
	header[4] = 0
	if hasAudio {
		header[4] |= 0x04
	}
	if hasVideo {
		header[4] |= 0x01
	}
	// data offset
	binary.BigEndian.PutUint32(header[5:9], 9)
	// previousTagSize 0 (always 0)
	binary.BigEndian.PutUint32(header[9:], 0)
	if n, err := w.Write(header); err != nil || n != len(header) {
		return fmt.Errorf("flv write header %d, error %v", n, err)
	}
	return nil
}

// WriteTag sends tag.
func (muxer *Muxer) WriteTag(w io.Writer, tag TagI) error {
	header := muxer.writeTagHeaderBuf[:]
	tagType, tagLen, tagTimestamp := tag.Info()
	writeTagHeader(header[:11], tagType, tagLen, tagTimestamp)
	if err := utils.WriteFull(w, header[:11]); err != nil {
		return err
	}
	if err := muxer.mux(w, tag); err != nil {
		return err
	}
	binary.BigEndian.PutUint32(header[11:15], uint32(tagLen)+11) // previousTagSizeN
	if err := utils.WriteFull(w, header[11:15]); err != nil {
		return err
	}
	return nil
}

func (muxer *Muxer) mux(w io.Writer, tag TagI) error {
	switch t := tag.(type) {
	case *AudioTag:
		return muxer.audioTag(w, t)
	case *VideoTag:
		return muxer.videoTag(w, t)
	case *ScriptTag:
		return scriptTag(w, t)
	default:
		return fmt.Errorf("flv muxer invalid tag %#v", tag)
	}
}

func (muxer *Muxer) audioTag(w io.Writer, tag *AudioTag) error {
	if tag.PacketType != SequenceHeader && tag.PacketType != AVPacket && tag.PacketType != EndOfSequence {
		return fmt.Errorf("flv muxer audio invalid packet type %d", tag.PacketType)
	}
	h := muxer.muxerAudioTagBuf[:]
	n := 1
	switch tag.SoundFormat {
	case AudioAAC:
		// 44k-SoundRate / 16-bit Samples / Stereo sound
		h[0] = (tag.SoundFormat << 4) | (3 << 2) | (1 << 1) | 1
		h[1] = tag.PacketType
		n++
	default:
		if tag.Channels != 0 && tag.Channels != 1 {
			return fmt.Errorf("flv muxer audio tag invalid Channels %d", tag.Channels)
		}
		if tag.SampleRate > 3 {
			return fmt.Errorf("flv muxer audio tag invalid SampleRate %d", tag.SampleRate)
		}
		if tag.BitPerSample != 0 && tag.BitPerSample != 1 {
			return fmt.Errorf("flv muxer audio tag invalid BitPerSample %d", tag.BitPerSample)
		}
		h[0] = (tag.SoundFormat << 4) | (tag.SampleRate&0x03)<<2 | (tag.BitPerSample&0x01)<<1 | (tag.Channels & 0x01)
	}
	if err := utils.WriteFull(w, h[:n]); err != nil {
		return err
	}
	return utils.WriteFull(w, tag.Data)
}
func (muxer *Muxer) videoTag(w io.Writer, tag *VideoTag) error {
	if tag.FrameType > 0x0f {
		return fmt.Errorf("flv muxer video invalid FrameType %d", tag.FrameType)
	}
	if tag.CodecID > 0x0f {
		return fmt.Errorf("flv muxer video invalid CodecID %d", tag.CodecID)
	}
	if tag.PacketType != SequenceHeader && tag.PacketType != AVPacket && tag.PacketType != EndOfSequence {
		return fmt.Errorf("flv muxer video invalid packet type %d", tag.PacketType)
	}
	h := muxer.muxerVideoTagBuf[:]
	n := 1
	h[0] = (tag.FrameType << 4) | tag.CodecID
	switch tag.CodecID {
	case VideoH264, VideoH265:
		compositionTime := tag.PTS - tag.DTS
		h[1] = tag.PacketType
		utils.BigEndianPutUint24(h[2:5], compositionTime)
		n += 4
	}
	if err := utils.WriteFull(w, h); err != nil {
		return err
	}
	return utils.WriteFull(w, tag.Data)
}

func scriptTag(w io.Writer, tag *ScriptTag) error {
	return utils.WriteFull(w, tag.Data)
}

func writeTagHeader(header []byte, tagType byte, size int, timestamp uint32) {
	_ = header[10] // bounds check
	header[0] = tagType
	// DataSize
	header[1] = byte(size >> 16)
	header[2] = byte(size >> 8)
	header[3] = byte(size)
	// Timestamp
	utils.BigEndianPutUint24(header[4:7], timestamp)
	header[7] = byte(timestamp >> 24) // extended timestamp
	// StreamID
	header[8] = 0
	header[9] = 0
	header[10] = 0
}
