package riff

import (
	"encoding/binary"
	"io"

	"github.com/foolishCDN/AV-spy/utils"
)

type Writer struct {
	w      io.Writer
	u32    [4]byte
	chunks []Chunk
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

func (s *Writer) WriteRIFFChunkHeader(fileFormat []byte, fileSize uint32) error {
	if err := utils.WriteFull(s.w, ChunkIDRIFF); err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(s.u32[:], fileSize)
	if err := utils.WriteFull(s.w, s.u32[:]); err != nil {
		return err
	}
	for len(fileFormat) < 4 {
		fileFormat = append(fileFormat, ' ')
	}
	if err := utils.WriteFull(s.w, fileFormat); err != nil {
		return err
	}
	return nil
}

func (s *Writer) WriteChunk(chunkID []byte, data []byte) error {
	for len(chunkID) < 4 {
		chunkID = append(chunkID, ' ')
	}
	if err := utils.WriteFull(s.w, chunkID); err != nil {
		return err
	}
	size := len(data)
	binary.LittleEndian.PutUint32(s.u32[:], uint32(size))
	if err := utils.WriteFull(s.w, s.u32[:]); err != nil {
		return err
	}
	if size%2 == 1 {
		data = append(data, byte(0))
	}
	return binary.Write(s.w, binary.LittleEndian, data)
}
