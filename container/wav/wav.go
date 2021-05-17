package wav

import (
	"github.com/sikasjc/AV-spy/encoding/riff"
)

var (
	RIFFTypeIDWAVE = []byte("WAVE")

	ChunkIDFMT    = []byte("fmt ")
	ChunkIDData   = []byte("data")
	ChunkIDSample = []byte("smpl")
	// ChunkIDWaveList contains a sequence of alternating silent chunks and data chunks.
	ChunkIDWaveList = []byte("wavl")
)

type Handler interface {
	OnPCM([]byte) error
	OnFormat(format *Format) error
	OnUnknownChunk(*riff.Chunk) error
}
