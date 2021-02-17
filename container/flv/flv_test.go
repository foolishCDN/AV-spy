package flv

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestMuxerAndDemuxer(t *testing.T) {
	f, err := os.Open("test.flv")
	if err != nil {
		t.Fatalf("open flv file err, %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	tmpFile, err := os.Create("test_copy.flv")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	demuxer := new(Demuxer)
	muxer := new(Muxer)
	header, err := demuxer.ReadHeader(f)
	if err != nil {
		t.Fatalf("read header err, %v", err)
	}

	if err := muxer.WriteHeader(tmpFile, header.HasAudio, header.HasVideo); err != nil {
		t.Fatal(err)
	}
	var maxSize int
	for {
		tag, err := demuxer.ReadTag(f)
		if err != nil {
			if err != io.EOF {
				t.Fatalf("read tag err, %v", err)
			} else {
				break
			}
		}
		d := tag.Len()
		if d > maxSize {
			maxSize = d
		}
		if err := muxer.WriteTag(tmpFile, tag); err != nil {
			t.Fatal(err)
		}
	}
	fmt.Println(maxSize)
}

func TestParser(t *testing.T) {
	r, err := os.Open("test.flv")
	if err != nil {
		t.Fatalf("open flv file err, %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	parser := NewParser(func(tag TagI) error {
		return nil
	})
	for {
		var b = make([]byte, 1024)
		n, err := r.Read(b)
		if err != nil {
			if err != io.EOF {
				t.Fatalf("read tag err, %v", err)
			} else {
				break
			}
		}
		if err := parser.Input(b[:n]); err != nil {
			t.Fatal(err)
		}
	}
}
