package main

import (
	"bytes"
	"fmt"
	"github.com/foolishCDN/AV-spy/codec"
	"github.com/foolishCDN/AV-spy/container/flv"
	"github.com/foolishCDN/AV-spy/encoding/amf"
	"github.com/foolishCDN/AV-spy/formatter"
	"github.com/sikasjc/pretty"
	"io"
)

var defaultVideoTemplate = formatter.NewTemplate("$stream_type:%6s $stream_id:%7d $pts:%7d $dts:%7d $size:%7d $frame_type $codec_id")
var defaultAudioTemplate = formatter.NewTemplate("$stream_type:%6s $stream_id:%7d $pts:%7d $dts:%7d $size:%7d $sound_format $channels $sound_size $sample_rate")
var defaultScriptTemplate = formatter.NewTemplate("$stream_type:%6s $stream_id:%7d $pts:%7d $dts:%7d $size:%7d")

var csvVideoTemplate = formatter.NewTemplate("$stream_type,$stream_id:%d,$pts:%d,$dts:%d,$size:%d,$frame_type,$codec_id")
var csvAudioTemplate = formatter.NewTemplate("$stream_type,$stream_id:%d,$pts:%d,$dts:%d,$size:%d,$sound_format,$channels,$sound_size,$sample_rate")
var csvScriptTemplate = formatter.NewTemplate("$stream_type,$stream_id:%d,$pts:%d,$dts:%d,$size:%d")

type FlvParser struct {
	titleDone       bool
	videoFormatter  formatter.Formatter
	audioFormatter  formatter.Formatter
	scriptFormatter formatter.Formatter
}

func (p *FlvParser) Println(tag flv.TagI) {
	if !p.titleDone {
		p.titleDone = true
		fmt.Printf("%16s %7s %7s %7s\n", "StreamID", "PTS", "DTS", "Size")
	}
	switch t := tag.(type) {
	case *flv.AudioTag:
		fmt.Println(p.audioFormatter.Format(t.ToVars()))
	case *flv.VideoTag:
		fmt.Println(p.videoFormatter.Format(t.ToVars()))
	case *flv.ScriptTag:
		fmt.Println(p.scriptFormatter.Format(t.ToVars()))
	}
}

func (p *FlvParser) OnHeader(header *flv.Header) {
	if !(showHeader || showAll) {
		return
	}
	fmt.Println("---------- FLV Header ----------")
	fmt.Printf("Version: %d\n", header.Version)
	fmt.Printf("HasVideo: %t\n", header.HasVideo)
	fmt.Printf("HasAudio: %t\n", header.HasAudio)
	fmt.Printf("HeaderSize: %d\n", header.DataOffset)
	fmt.Println("------------------------------")
}

func (p *FlvParser) OnPacket(tag flv.TagI) error {
	switch t := tag.(type) {
	case *flv.AudioTag:
		if t.SoundFormat == flv.AAC && t.PacketType == flv.SequenceHeader {
			p.OnAAC(t)
		}
	case *flv.VideoTag:
		if t.PacketType == flv.SequenceHeader {
			p.OnAVC(t)
		}
	case *flv.ScriptTag:
		decoder := amf.NewDecoder(amf.Version0)
		buf := bytes.NewBuffer(t.Bytes)
		got, err := decoder.DecodeBatch(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
		}
		if showMetaData || showAll {
			fmt.Println("---------- MetaData ----------")
			pretty.Println(got)
			fmt.Println("------------------------------")
		}
	}
	if !(showPacket || showAll) {
		return nil
	}
	p.Println(tag)
	return nil
}

func (p *FlvParser) OnAAC(t *flv.AudioTag) error {
	if !(showExtraData || showAll) {
		return nil
	}
	aac := new(codec.AACAudioSpecificConfig)
	if err := aac.Read(t.Bytes); err != nil {
		return err
	}
	fmt.Println("-- sequence header of audio --")
	pretty.Println(aac)
	fmt.Println("------------------------------")
	return nil
}

func (p *FlvParser) OnAVC(t *flv.VideoTag) error {
	if !(showExtraData || showAll) {
		return nil
	}
	avc := new(codec.AVCDecoderConfigurationRecord)
	if err := avc.Read(t.Bytes); err != nil {
		return err
	}
	fmt.Println("-- sequence header of video --")
	pretty.Println(avc)
	fmt.Println("------------------------------")
	return nil
}

func NewFlvParser(format string) (*FlvParser, error) {
	p := &FlvParser{}
	switch format {
	case DefaultFormat:
		p.videoFormatter = defaultVideoTemplate
		p.audioFormatter = defaultAudioTemplate
		p.scriptFormatter = defaultScriptTemplate
	case "csv":
		p.videoFormatter = csvVideoTemplate
		p.audioFormatter = csvAudioTemplate
		p.scriptFormatter = csvScriptTemplate
	default:
		return nil, fmt.Errorf("format %q not supported", format)
	}
	return p, nil
}
