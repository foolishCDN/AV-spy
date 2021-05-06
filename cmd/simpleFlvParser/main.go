package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gobs/pretty"

	"github.com/sikasjc/AV-spy/codec"

	"github.com/sikasjc/AV-spy/container/flv"
	"github.com/sikasjc/AV-spy/encoding/amf"
)

var showExtraData = flag.Bool("show_extradata", false, "will show codec extradata(sequence header)")
var showPackets = flag.Bool("show_packets", false, "will show packets info")
var showMetaData = flag.Bool("show_metadata", true, "will show meta data")

func main() {
	log.SetFlags(0)
	log.SetPrefix("simpleFlvParser: ")
	flag.Usage = usage
	flag.Parse()
	printer := pretty.Pretty{Compact: true, Out: os.Stdout}

	args := flag.Args()
	if args == nil || len(args) < 1 {
		log.Println("no input file")
		usage()
		return
	}
	path := args[0]
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
			if !*showPackets {
				continue
			}
			fmt.Printf("{ AUDIO} %d %d %s %s %s %s\n",
				t.StreamID, t.PTS, flv.AudioSoundFormatMap[t.SoundFormat], flv.AudioSoundTypeMap[t.Channels], flv.AudioSoundSizeMap[t.BitPerSample], flv.AudioSoundRateMap[t.SampleRate])
		case *flv.VideoTag:
			if t.PacketType == flv.SequenceHeader && *showExtraData {
				avc := new(codec.AVCDecoderConfigurationRecord)
				if err := avc.Read(t.Data); err != nil {
					log.Fatal(err)
				}
				fmt.Println("-- sequence header of video --")
				printer.Println(avc)
				fmt.Println("------------------------------")
				continue
			}
			if !*showPackets {
				continue
			}
			fmt.Printf("{ VIDEO} %d %d %s %s\n",
				t.StreamID, t.PTS, flv.VideoFormatTypeMap[t.FrameType], flv.VideoCodecIDMap[t.CodecID])

		case *flv.ScriptTag:
			decoder := amf.NewDecoder(amf.Version0)
			buf := bytes.NewBuffer(t.Data)
			got, err := decoder.DecodeBatch(buf)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
			}
			if *showMetaData {
				fmt.Println("---------- MetaData ----------")
				printer.Println(got)
				fmt.Println("------------------------------")
			}
			if !*showPackets {
				continue
			}
			fmt.Printf("{SCRIPT} %d %d\n", t.StreamID, t.PTS)
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n\t"+
		"simpleFlvParser [options] input.flv\n\t"+
		"simpleFlvParser [options] http://path/to/input.flv\n\n"+
		"Options:\n")
	flag.PrintDefaults()
}

func parseFilePath(path string) (io.ReadCloser, error) {
	if isValidURL(path) {
		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("new request err: %v", err)
		}
		client := &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 10 * time.Second,
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request err: %v", err)
		}
		return resp.Body, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file err: %v", err)
	}
	return f, nil
}

func isValidURL(path string) bool {
	u, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	if u.Scheme == "" || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
