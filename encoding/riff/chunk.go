package riff

var ChunkIDRIFF = []byte{'R', 'I', 'F', 'F'}

const (
	StateReadRIFFChunkHeader = iota
	StateReadChunkHeader
	StateReadChunkData
	StateReadDone
)

// Chunk must be word aligned.
// Please use Size, not len(Data)
type Chunk struct {
	ID   []byte
	Size uint32
	Data []byte
}

// RIFFChunkHeader
// ID is 'RIFF'
// Can contain multiple Chunk
type RIFFChunkHeader struct {
	Size       uint32
	FormatType []byte
}
