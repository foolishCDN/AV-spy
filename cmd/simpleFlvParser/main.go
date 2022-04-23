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

	"github.com/foolishCDN/AV-spy/codec"
	"github.com/foolishCDN/AV-spy/container/flv"
	"github.com/foolishCDN/AV-spy/encoding/amf"

	"github.com/gobs/pretty"
)

var showExtraData = flag.Bool("show_extradata", false, "will show codec extradata(sequence header)")
var showPackets = flag.Bool("show_packets", false, "will show packets info")
var showMetaData = flag.Bool("show_metadata", false, "will show meta data")
var showHeader = flag.Bool("show_header", false, "will show flv file header")
var showAll = flag.Bool("show", false, "will show all message")

var num = flag.Int("n", 0, "show `n` tags, default: no limit")

var printer = pretty.Pretty{Compact: true, Out: os.Stdout}

func OnHeader(header *flv.Header) {
	if !(*showHeader || *showAll) {
		return
	}
	fmt.Println("---------- FLV Header ----------")
	fmt.Printf("Version: %d\n", header.Version)
	fmt.Printf("HasVideo: %t\n", header.HasVideo)
	fmt.Printf("HasAudio: %t\n", header.HasAudio)
	fmt.Printf("HeaderSize: %d\n", header.DataOffset)
	fmt.Println("------------------------------")
}

func OnAAC(t *flv.AudioTag) {
	if !(*showExtraData || *showAll) {
		return
	}
	aac := new(codec.AACAudioSpecificConfig)
	if err := aac.Read(t.Bytes); err != nil {
		log.Fatal(err)
	}
	fmt.Println("-- sequence header of audio --")
	printer.Println(aac)
	fmt.Println("------------------------------")
}

func OnAVC(t *flv.VideoTag) {
	if !(*showExtraData || *showAll) {
		return
	}
	avc := new(codec.AVCDecoderConfigurationRecord)
	if err := avc.Read(t.Bytes); err != nil {
		log.Fatal(err)
	}
	fmt.Println("-- sequence header of video --")
	printer.Println(avc)
	fmt.Println("------------------------------")
}

func OnAudio(t *flv.AudioTag) {
	if !(*showPackets || *showAll) {
		return
	}
	label := "{ AUDIO}"
	if t.SoundFormat == flv.AAC && t.PacketType == flv.SequenceHeader {
		label = "{   AAC}"
		OnAAC(t)
	}
	fmt.Printf("%s %7d %7d %7d %7d %s %s %s %s\n",
		label, t.StreamID, t.PTS, t.PTS, len(t.Data()), t.SoundFormat.String(), t.Channels.String(), t.BitPerSample.String(), t.SampleRate.String())
}

func OnVideo(t *flv.VideoTag) {
	if !(*showPackets || *showAll) {
		return
	}
	label := "{ VIDEO}"
	if t.PacketType == flv.SequenceHeader {
		label = "{   AVC}"
		if t.CodecID == flv.H265 {
			label = "{  HAVC}"
		}
		OnAVC(t)
	}
	fmt.Printf("%s %7d %7d %7d %7d %s %s\n",
		label, t.StreamID, t.PTS, t.DTS, len(t.Data()), t.FrameType.String(), t.CodecID.String())
}

func OnScript(t *flv.ScriptTag) {
	decoder := amf.NewDecoder(amf.Version0)
	buf := bytes.NewBuffer(t.Bytes)
	got, err := decoder.DecodeBatch(buf)
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
	}
	if *showMetaData || *showAll {
		fmt.Println("---------- MetaData ----------")
		printer.Println(got)
		fmt.Println("------------------------------")
	}
	if !(*showPackets || *showAll) {
		return
	}
	fmt.Printf("%16s %7s %7s %7s\n", "StreamID", "PTS", "DTS", "Size")
	fmt.Printf("{SCRIPT} %7d %7d %7d\n", t.StreamID, t.PTS, len(t.Data()))
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("simpleFlvParser: ")
	flag.Usage = usage
	flag.Parse()
	if !(*showPackets || *showHeader || *showExtraData || *showMetaData || *showAll) {
		flag.PrintDefaults()
		log.Fatal("Please set options !!!")
	}

	args := flag.Args()
	if args == nil || len(args) < 1 {
		usage()
		log.Fatal("No input file !!!")
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
	OnHeader(header)
	count := 0
	for {
		tag, err := demuxer.ReadTag(r)
		if err != nil {
			if err == io.EOF {
				return
			} else {
				log.Fatal(err)
			}
		}
		count++
		switch t := tag.(type) {
		case *flv.AudioTag:
			OnAudio(t)
		case *flv.VideoTag:
			OnVideo(t)
		case *flv.ScriptTag:
			OnScript(t)
		}
		if *num > 0 && count > *num {
			break
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
