package wav

import (
	"github.com/sikasjc/AV-spy/encoding/riff"
)

var (
	RIFFTypeIDWAVE = []byte("WAVE")

	ChunkIDFMT    = []byte("fmt ")
	ChunkIDDATA   = []byte("data")
	ChunkIDSample = []byte("smpl")
)

type Handler interface {
	OnPCM([]byte) error
	OnFormat(format *Format) error
	OnUnknownChunk(*riff.Chunk) error
}
