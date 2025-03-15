package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/foolishCDN/AV-spy/codec"
	"github.com/foolishCDN/AV-spy/codec/avc"
	"github.com/foolishCDN/AV-spy/codec/hevc"
	"github.com/foolishCDN/AV-spy/container/flv"
	"github.com/foolishCDN/AV-spy/encoding/amf"
	"github.com/foolishCDN/AV-spy/formatter"
	"github.com/foolishCDN/AV-spy/summary"
	"github.com/foolishCDN/AV-spy/utils"
	"github.com/sikasjc/pretty"
	"github.com/sirupsen/logrus"
)

var defaultVideoTemplate = formatter.NewTemplate("$stream_type:%6s $stream_id:%7d $pts:%7d $dts:%7d $size:%7d $frame_type $codec_id $nalu_types")
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

	videoCounter summary.Counter
	audioCounter summary.Counter
	sps          codec.SPS
	codec        string
}

func (p *FlvParser) Println(tag flv.TagI) {
	if !p.titleDone {
		p.titleDone = true
		fmt.Printf("%16s %7s %7s %7s\n", "StreamID", "PTS", "DTS", "Size")
	}
	switch t := tag.(type) {
	case *flv.AudioTag:
		if showPacket {
			fmt.Println(p.audioFormatter.Format(t.ToVars()))
		}
	case *flv.VideoTag:
		payloadType, payloadSize, payload := t.SEI()
		if showSEI {
			if payloadType != 0 {
				fmt.Println("------------- SEI ------------")
				pretty.Println(fmt.Sprint("payload type: ", payloadType))
				pretty.Println(fmt.Sprint("payload size: ", payloadSize))
				switch seiFormat {
				case seiFormatHex:
					pretty.Println(payload)
				case seiFormatByte:
					pretty.DefaultPrinter.CompactArray = true
					pretty.Println(payload)
				case seiFormatString:
					pretty.Println(string(payload))
				}
				fmt.Println("------------------------------")
				fmt.Println(p.videoFormatter.Format(t.ToVars()))
			}
		}
		if showPacket {
			fmt.Println(p.videoFormatter.Format(t.ToVars()))
		}
	case *flv.ScriptTag:
		if showPacket {
			fmt.Println(p.scriptFormatter.Format(t.ToVars()))
		}
	}
}

func (p *FlvParser) Summary() {
	v := p.videoCounter
	a := p.audioCounter
	fmt.Println("\nSummary:")
	fmt.Printf("  Running time: %v\n", v.Duration())
	if v.Total > 0 {
		fmt.Println("  video:")
		if p.sps != nil {
			if p.sps.FPS() > 0 {
				fmt.Printf("    resolution: %dx%d, codec: %s, fps: %.2f (from sps)\n",
					p.sps.Width(), p.sps.Height(), p.codec, p.sps.FPS())
			} else {
				fmt.Printf("    resolution: %dx%d, codec: %s\n",
					p.sps.Width(), p.sps.Height(), p.codec)
			}
		}
		fmt.Printf("    count/timestamp: %d/%d, fps: %.2f, real fps: %0.2f, gap: %d, rewind: %d, duplicate: %d\n",
			v.Total, v.TimestampDuration(), v.Rate(), v.RealRate(), v.MaxGap, v.MaxRewind, v.Duplicate)
		cacheTimestampDuration := v.CacheTimestampDuration()
		cacheDuration := v.CacheDuration()
		estimatedCacheFps := v.EstimatedCacheFps()
		if cacheTimestampDuration == 0 {
			fmt.Printf("    Estimated cache: %d(not yet over) was send within %v\n", v.TimestampDuration(), v.Duration())
		} else {
			fmt.Printf("    Estimated cache: %d was send within %v, estimated fps: %0.2f\n", cacheTimestampDuration, cacheDuration, estimatedCacheFps)
		}
	}
	if a.Total > 0 {
		fmt.Println("  audio:")
		fmt.Printf("    count/timestamp: %d/%d, pps: %.2f, real pps: %0.2f, gap: %d, rewind: %d, duplicate: %d\n",
			a.Total, a.TimestampDuration(), a.Rate(), a.RealRate(), a.MaxGap, a.MaxRewind, a.Duplicate)
		cacheTimestampDuration := a.CacheTimestampDuration()
		cacheDuration := a.CacheDuration()
		estimatedCacheFps := a.EstimatedCacheFps()
		if cacheTimestampDuration == 0 {
			fmt.Printf("    Estimated cache: %d(not yet over) was send within %v\n", a.TimestampDuration(), a.Duration())
		} else {
			fmt.Printf("    Estimated cache: %d was send within %v, estimated fps: %0.2f\n", cacheTimestampDuration, cacheDuration, estimatedCacheFps)
		}
	}

}

func (p *FlvParser) OnHeader(header *flv.Header) {
	if !(showHeader) {
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
			if err := p.OnAAC(t); err != nil {
				logrus.WithField("error", err).Error("parse sequence header of audio AACAudioSpecificConfig failed")
			}
		} else {
			p.audioCounter.Count(int(t.PTS))
		}
	case *flv.VideoTag:
		if t.PacketType == flv.SequenceHeader {
			switch t.CodecID {
			case flv.H264:
				p.codec = "avc"
				if err := p.OnAVC(t); err != nil {
					logrus.WithField("error", err).Error("parse sequence header of video AVCDecoderConfigurationRecord failed")
				}
			case flv.H265:
				p.codec = "hevc"
				if err := p.OnHEVC(t); err != nil {
					logrus.WithField("error", err).Error("parse sequence header of video HEVCDecoderConfigurationRecord failed")
				}
			default:
				logrus.WithField("codec_id", t.CodecID).Warn("unknown sequence header type of video")
			}
		} else {
			p.videoCounter.Count(int(t.DTS))
		}
	case *flv.ScriptTag:
		decoder := amf.NewDecoder(amf.Version0)
		buf := bytes.NewBuffer(t.Bytes)
		got, err := decoder.DecodeBatch(buf)
		if err != nil {
			if err != io.EOF {
				logrus.WithField("error", err).Error("parse script tag failed")
			}
		}
		if showMetaData {
			fmt.Println("---------- MetaData ----------")
			pretty.Println(got)
			fmt.Println("------------------------------")
		}
	}
	if !(showPacket || showSEI) {
		return nil
	}
	p.Println(tag)
	return nil
}

func (p *FlvParser) OnAAC(t *flv.AudioTag) error {
	if !(showExtraData) {
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
	if !(showExtraData) {
		return nil
	}
	decoderConfigurationRecord := new(avc.AVCDecoderConfigurationRecord)
	if err := decoderConfigurationRecord.Read(t.Bytes); err != nil {
		return err
	}
	t.NALUs = append(t.NALUs, decoderConfigurationRecord.SPS...)
	t.NALUs = append(t.NALUs, decoderConfigurationRecord.PPS...)
	fmt.Println("-- sequence header of video --")
	pretty.Println(decoderConfigurationRecord)
	reader := utils.NewBitReader(decoderConfigurationRecord.SPS[0])
	avc.ParseNALUHeader(reader)
	sps, err := avc.ParseSPS(reader)
	if err != nil {
		logrus.Debugf("parse sps failed, the hex string of decoderConfigurationRecord is %s", hex.EncodeToString(t.Bytes))
	}
	p.sps = sps
	fmt.Println("-- From SPS --")
	fmt.Printf("resolution: %dx%d\n", sps.Width(), sps.Height())
	fmt.Printf("fps: %.2f (It's not mandatory)\n", sps.FPS())
	fmt.Println("------------------------------")
	return nil
}

func (p *FlvParser) OnHEVC(t *flv.VideoTag) error {
	if !(showExtraData) {
		return nil
	}
	decoderConfigurationRecord := new(hevc.HEVCDecoderConfigurationRecord)
	if err := decoderConfigurationRecord.Read(t.Bytes); err != nil {
		return err
	}
	fmt.Println("-- sequence header of video --")
	pretty.Println(decoderConfigurationRecord)
	for _, ps := range decoderConfigurationRecord.NALUs {
		t.NALUs = append(t.NALUs, ps.NALUs...)
		if ps.NALUnitType != hevc.NalSPS {
			continue
		}
		reader := utils.NewBitReader(ps.NALUs[0])
		hevc.ParseNALUHeader(reader)
		sps, err := hevc.ParseSPS(reader)
		if err != nil {
			logrus.Debugf("parse sps failed, the hex string of decoderConfigurationRecord is %s", hex.EncodeToString(t.Bytes))
		}
		p.sps = sps
		fmt.Println("-- From SPS --")
		fmt.Printf("resolution: %dx%d\n", sps.Width(), sps.Height())
		fmt.Printf("fps: %.2f (It's not mandatory)\n", sps.FPS())
		fmt.Println("------------------------------")
	}
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
