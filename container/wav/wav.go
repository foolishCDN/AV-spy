package wav

import (
	"github.com/sikasjc/AV-spy/encoding/riff"
)

var (
	RIFFTypeIDWAVE = []byte("WAVE")

	ChunkIDFMT  = []byte("fmt ")
	ChunkIDData = []byte("data")
	// ChunkIDSample allows a MIDI sampler to use the Wave file as a collection of samples
	ChunkIDSample = []byte("smpl")
	// ChunkIDWaveList contains a sequence of alternating silent chunks and data chunks
	ChunkIDWaveList = []byte("wavl")
)

type Handler interface {
	OnPCM([]byte) error
	OnFormat(*Format) error
	OnMIDISample(sample *MIDISample) error
	OnUnknownChunk(*riff.Chunk) error
}
