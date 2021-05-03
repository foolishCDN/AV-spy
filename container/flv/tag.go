package flv

// Header FLV file header
type Header struct {
	Version    byte
	HasAudio   bool
	HasVideo   bool
	DataOffset uint32
}

// TagI ...
type TagI interface {
	Info() (tagType byte, len int, timestamp uint32)
}

// AudioTag ...
type AudioTag struct {
	SoundFormat  byte
	SampleRate   byte // 0-5.5kHz, 1-11kHz, 2-22kHz, 3-44kHz
	BitPerSample byte // 0-8bit, 1-16bit
	Channels     byte // 0-mono, 1-stereo

	PTS        uint32
	StreamID   uint32
	PacketType byte // 0-AAC sequence header, 1-AAC raw if SoundFormat=10
	Data       []byte
}

// Info ...
func (tag *AudioTag) Info() (tagType byte, size int, timestamp uint32) {
	tagType = TagAudio
	timestamp = tag.PTS
	size = len(tag.Data) + 1
	if tag.SoundFormat == AudioAAC {
		size++
	}
	return tagType, size, timestamp
}

// VideoTag ...
type VideoTag struct {
	FrameType byte
	CodecID   byte

	DTS        uint32
	PTS        uint32
	StreamID   uint32
	PacketType byte // 0-AVC sequence header, 1-AVC NALU, 2-AVC end of sequence if CodecID=7
	Data       []byte
}

// Info ...
func (tag *VideoTag) Info() (tagType byte, size int, timestamp uint32) {
	tagType = TagVideo
	timestamp = tag.DTS
	size = len(tag.Data) + 1
	if tag.CodecID == VideoH264 || tag.CodecID == VideoH265 {
		size += 4
	}
	return tagType, size, timestamp
}

// ScriptTag ...
type ScriptTag struct {
	PTS      uint32
	StreamID uint32
	Data     []byte
}

// Info ...
func (tag *ScriptTag) Info() (tagType byte, size int, timestamp uint32) {
	tagType = TagScript
	timestamp = tag.PTS
	size = len(tag.Data)
	return tagType, size, timestamp
}
