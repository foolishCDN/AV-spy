package flv

var Signature = []byte("FLV")

type TagType byte

// Tag Type
const (
	TagAudio  TagType = 8  // Audio
	TagVideo  TagType = 9  // Video
	TagScript TagType = 18 // Script
)

// AAC/AVC sequence header
const (
	SequenceHeader byte = iota
	AVPacket
	EndOfSequence
)

type Format byte

// Video FormatType
const (
	KeyFrame Format = iota + 1
	InterFrame
	DisposableInterFrame // H263 only
	GeneratedKeyFrame    // reserved for server use only
	InfoFrame
)

type CodecID byte

// Video CodecID
const (
	JPEG CodecID = iota + 1
	H263
	ScreenVideo
	On2VP6
	On2VP6WithAlpha
	ScreenVideoV2
	H264
	H265 = 12 // H265 https://github.com/CDN-Union/H265
)

type SoundFormat byte

// Audio SoundFormat
const (
	LinearPCM SoundFormat = iota
	ADPCM
	MP3
	PCM
	Nellymoser16KHzMono
	Nellymoser8KHzMono
	Nellymoser
	G711A
	G711U
	_ // reserved
	AAC
	Speex
	MP38KHz
	DeviceSpecificSound
)

type SoundRate byte

// Audio SoundRate
const (
	_5KHz  SoundRate = iota // 5KHz
	_11KHz                  // 11KHz
	_22KHz                  // 22KHz
	_44KHz                  // 44KHz
)

type SoundSize byte

// Audio SoundSize
const (
	_8bit  SoundSize = iota // 8bit
	_16bit                  // 16bit
)

type SoundType byte

// Audio SoundType
const (
	Mono SoundType = iota
	Stereo
)
