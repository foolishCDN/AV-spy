package flv

var Signature = []byte("FLV")

// Tag Type
const (
	TagAudio  byte = 8
	TagVideo  byte = 9
	TagScript byte = 18
)

// enum_info:no
// AAC/AVC sequence header
const (
	SequenceHeader byte = iota
	AVPacket
	EndOfSequence
)

// Video FormatType
const (
	VideoKeyFrame byte = iota + 1
	VideoInterFrame
	VideoDisposableInterFrame // H263 only
	VideoGeneratedKeyFrame    // reserved for server use only
	VideoInfoFrame
)

// Video CodecID
const (
	VideoJPEG byte = iota + 1
	VideoH263
	VideoScreenVideo
	VideoOn2VP6
	VideoOn2VP6WithAlpha
	VideoScreenVideoV2
	VideoH264
	VideoH265 = 12 // https://github.com/CDN-Union/H265
)

// Audio SoundFormat
const (
	AudioLinearPCM byte = iota
	AudioADPCM
	AudioMP3
	AudioPCM
	AudioNellymoser16KHzMono
	AudioNellymoser8KHzMono
	AudioNellymoser
	AudioG711A
	AudioG711U
	_ // reserved
	AudioAAC
	AudioSpeex
	AudioMP38KHz
	AudioDeviceSpecificSound
)

// Audio SoundRate
const (
	Audio5KHz byte = iota
	Audio11KHz
	Audio22KHz
	Audio44KHz // AAC
)

// Audio SoundSize
const (
	Audio8bit byte = iota
	Audio16bit
)

// Audio SoundType
const (
	AudioMono byte = iota
	AudioStereo
)
