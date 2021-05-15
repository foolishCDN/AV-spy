package riff

import (
	"io"
	"os"
	"testing"

	"github.com/udhos/equalfile"
)

func TestParserAndWriter(t *testing.T) {
	f, err := os.Open("test.wav")
	if err != nil {
		t.Fatalf("open wav file err, %v", err)
	}
	defer f.Close()

	fCopy, err := os.Create("test_copy.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer fCopy.Close()

	writer := NewWriter(fCopy)
	parser := &Parser{
		OnChunk: func(chunk *Chunk) error {
			return writer.WriteChunk(chunk.ID, chunk.Data)
		},
		OnRIFFChunkHeader: func(header *RIFFChunkHeader) error {
			return writer.WriteRIFFChunkHeader(header.FormatType, header.Size)
		},
	}

	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		if err := parser.Input(buf[:n]); err != nil {
			t.Fatal(err)
		}
	}
	cmp := equalfile.New(nil, equalfile.Options{})
	equal, err := cmp.CompareReader(f, fCopy)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Error("file content is not equal")
	}
}
