// Package mp3
//
// I think mp3 should be a codec,
// see https://developer.mozilla.org/en-US/docs/Web/Media/Formats/Containers#common_container_formats
// "A good example of this is the MP3 audio file,
// which is in fact an MPEG-1 container with a single audio track encoded using MPEG-1 Audio Layer III encoding."
//
// mp3 = (MPEG-1 Audio Layer III)
//
package mp3

import "errors"

var (
	ErrInvalidSyncWord               = errors.New("invalid sync word")
	ErrInvalidVersionID              = errors.New("invalid version id")
	ErrInvalidSamplingFrequencyIndex = errors.New("invalid sampling frequency index")
	ErrInvalidBitrateIndex           = errors.New("invalid bitrate index")
	ErrInvalidLayer                  = errors.New("invalid layer")
)

var (
	TagVer1Identifier = []byte{'T', 'A', 'G'}
	TagVer2Identifier = []byte{'I', 'D'}

	// SamplingFrequencyMap
	// SamplingFrequencyIndex [ Version ]: value (Hz)
	SamplingFrequencyMap = [3][4]int{
		0: {11025, 0, 22050, 44100},
		1: {12000, 0, 24000, 48000},
		2: {8000, 0, 16000, 32000},
	}
	// BitrateMap
	// Version [ Layer [ Bitrate ]]: value (kbps)
	BitrateMap = [4][4][16]int{
		MPEGVersion1: {
			Layer1: {0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 0},
			Layer2: {0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, 0},
			Layer3: {0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0},
		},
		// version 2 == version2.5, layer2 == layer3
		MPEGVersion2: {
			Layer1: {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 0},
			Layer2: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
			Layer3: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
		},
		MPEGVersion25: {
			Layer1: {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 0},
			Layer2: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
			Layer3: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
		},
	}

	// FrameSizeMap
	// Layer: [ Version ]: value (bytes)
	FrameSizeMap = [4][4]int{
		Layer1: {384, 384, 384},
		Layer2: {1152, 1152, 1152},
		Layer3: {MPEGVersion1: 1152, MPEGVersion2: 576, MPEGVersion25: 576},
	}
)

const (
	MPEGVersion25 byte = iota // version 2.5
	_                         // reserved
	MPEGVersion2
	MPEGVersion1
)

const (
	_ byte = iota
	Layer3
	Layer2
	Layer1
)

const (
	ModeStereo byte = iota
	ModeJointStereo
	ModeDual
	ModeSingle
)
