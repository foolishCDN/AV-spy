package flv

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

// VideoTag ...
type VideoTag struct {
	FrameType FrameType
	CodecID   CodecID

	DTS        uint32
	PTS        uint32
	StreamID   uint32
	PacketType byte // 0-AVC sequence header, 1-AVC NALU, 2-AVC end of sequence if CodecID=7
	Bytes      []byte
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
