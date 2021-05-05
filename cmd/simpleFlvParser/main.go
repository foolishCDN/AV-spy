package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/sikasjc/AV-spy/container/flv"
	"github.com/sikasjc/AV-spy/encoding/amf"
)

func main() {
	args := os.Args
	if args == nil || len(args) < 2 {
		usage()
		return
	}
	path := args[1]
	r, err := parseFilePath(path)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	demuxer := new(flv.Demuxer)
	header, err := demuxer.ReadHeader(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("---------- FLV Header ----------")
	fmt.Printf("Version: %d\n", header.Version)
	fmt.Printf("HasVideo: %t\n", header.HasVideo)
	fmt.Printf("HasAudio: %t\n", header.HasAudio)
	fmt.Printf("HeaderSize: %d\n", header.DataOffset)
	fmt.Println("------------------------------")
	for {
		tag, err := demuxer.ReadTag(r)
		if err != nil {
			if err == io.EOF {
				return
			} else {
				log.Fatal(err)
			}
		}
		switch t := tag.(type) {
		case *flv.AudioTag:
			fmt.Printf("{ AUDIO} %d %d %s %s %s %s\n",
				t.StreamID, t.PTS, flv.AudioSoundFormatMap[t.SoundFormat], flv.AudioSoundTypeMap[t.Channels], flv.AudioSoundSizeMap[t.BitPerSample], flv.AudioSoundRateMap[t.SampleRate])
		case *flv.VideoTag:
			fmt.Printf("{ VIDEO} %d %d %s %s\n",
				t.StreamID, t.PTS, flv.VideoFormatTypeMap[t.FrameType], flv.VideoCodecIDMap[t.CodecID])
		case *flv.ScriptTag:
			fmt.Printf("{SCRIPT} %d %d\n", t.StreamID, t.PTS)
			decoder := amf.NewDecoder(amf.Version0)
			buf := bytes.NewBuffer(t.Data)
			got, err := decoder.DecodeBatch(buf)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
			}
			fmt.Println("---------- MetaData ----------")
			spew.Dump(got)
			fmt.Println("------------------------------")
		}
	}
}

func usage() {
	fmt.Println("usage:")
	fmt.Println("simpleFlvParser input.flv")
	fmt.Println("simpleFlvParser http://path/to/input.flv")
}

func parseFilePath(path string) (io.ReadCloser, error) {
	if strings.Contains(path, "http://") {
		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}
		client := &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 10 * time.Second,
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}
