package mp3

import "io"

// FrameHeader is 4 bytes long and contains sync word to indicate the start of frame.
// Reference: http://www.goat.cz/index.php?path=MP3_MP3ProfiInfo
type FrameHeader struct {
	VersionID byte
	Layer     byte
	// Frames may have some form of check sum - CRC check. CRC is 16 bit long, and
	// if exists, it follows frame header. And then comes audio data.
	IsProtected            bool
	BitrateIndex           byte
	SamplingFrequencyIndex byte
	IsPadded               bool
	PrivateBit             byte
	Mode                   byte
	ModeExtension          byte
	IsCopyRight            bool
	IsCopy                 bool
	Emphasis               byte

	// Frame size is the number of samples contained in a frame.
	FrameSize int

	// Frame length is the length of a frame when compressed.
	// Frame length may change from frame to frame due to padding or bitrate switching.
	FrameLength int

	SideInfoSize int

	UseIntensityStereo bool
	UseMSStereo        bool
}

func ReadFrameHeader(r io.Reader) (*FrameHeader, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	if buf[0] != 0xff && buf[1]&0xe0 != 0xe0 { // [1,11] sync word, should always set
		return nil, ErrInvalidSyncWord
	}
	header := &FrameHeader{
		VersionID:              buf[1] & 0x18,    // [12, 13]
		Layer:                  buf[1] & 0x06,    // [14, 15]
		IsProtected:            buf[1]&0x01 == 0, // [15]
		BitrateIndex:           buf[2] & 0xf0,    // [17, 20]
		SamplingFrequencyIndex: buf[2] & 0x0c,    // [21, 22]
		IsPadded:               buf[2]&0x02 == 1, // [23]
		PrivateBit:             buf[2] & 0x01,    // [24]
		Mode:                   buf[3] & 0xc0,    // [25, 26]
		ModeExtension:          buf[3] & 0x30,    // [27, 28]
		IsCopyRight:            buf[3]&0x08 == 1, // [29]
		IsCopy:                 buf[3]&0x04 == 0, // [30]
		Emphasis:               buf[3] & 0x03,    // [31, 32]
	}

	// validation
	if header.VersionID == 1 {
		return nil, ErrInvalidVersionID
	}
	if header.Layer == 0 {
		return nil, ErrInvalidLayer
	}
	if header.SamplingFrequencyIndex > 2 {
		return nil, ErrInvalidSamplingFrequencyIndex
	}
	if header.BitrateIndex == 0 || header.BitrateIndex == 15 {
		return nil, ErrInvalidBitrateIndex
	}

	// compute frame length
	freq := SamplingFrequencyMap[header.SamplingFrequencyIndex][header.VersionID]
	bitrate := BitrateMap[header.VersionID][header.Layer][header.BitrateIndex]
	if header.IsPadded {
		if header.VersionID == MPEGVersion1 {
			header.FrameLength = 144*bitrate*1000/freq + 4
		} else {
			header.FrameLength = 144*bitrate*1000/freq + 1
		}
	} else {
		header.FrameLength = 144 * bitrate * 1000
	}

	// compute side info size
	if header.VersionID == MPEGVersion1 {
		if header.Mode == ModeSingle {
			header.SideInfoSize = 17
		} else {
			header.SideInfoSize = 32
		}
	} else {
		if header.Mode == ModeSingle {
			header.SideInfoSize = 9
		} else {
			header.SideInfoSize = 17
		}
	}

	// compute frame size
	header.FrameSize = FrameSizeMap[header.Layer][header.VersionID]

	// intensity stereo and mid side (MS) stereo are on or off
	if header.Mode == ModeJointStereo {
		header.UseIntensityStereo = header.ModeExtension&0x1 == 1
		header.UseMSStereo = header.ModeExtension&0x02 == 1
	} else {
		header.UseMSStereo = false
		header.UseIntensityStereo = false
	}
	return header, nil
}
