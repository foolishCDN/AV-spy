package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/awesome-gocui/gocui"
	"github.com/fatih/color"
	"github.com/foolishCDN/AV-spy/codec"
	"github.com/foolishCDN/AV-spy/container/flv"
	"github.com/foolishCDN/AV-spy/encoding/amf"
)

func onTag(g *gocui.Gui, tag flv.TagI, w io.Writer) {
	switch t := tag.(type) {
	case *flv.ScriptTag:
		onScript(g, t, w)
	case *flv.AudioTag:
		onAudio(g, t, w)
	case *flv.VideoTag:
		onVideo(g, t, w)
	}
}

func onAAC(g *gocui.Gui, t *flv.AudioTag, w io.Writer) {
	aac := new(codec.AACAudioSpecificConfig)
	if err := aac.Read(t.Bytes); err != nil {
		showError(g, "Parse aac fail, err %v\n", err)
		return
	}
	if w != nil {
		submitEvent(func(gui *gocui.Gui) error {
			_, _ = fmt.Fprintf(w, color.RedString("AAC Audio Specific Config:\n"))
			prettyPrintTo(w, aac)
			_, _ = fmt.Fprintf(w, "\n")
			return nil
		})
	}
}

func onAVC(g *gocui.Gui, t *flv.VideoTag, w io.Writer) {
	avc := new(codec.AVCDecoderConfigurationRecord)
	if err := avc.Read(t.Bytes); err != nil {
		showError(g, "Parse avc fail, err %v\n", err)
		return
	}
	if w != nil {
		submitEvent(func(gui *gocui.Gui) error {
			_, _ = fmt.Fprintf(w, color.RedString("AVC Decoder Configuration Record:\n"))
			prettyPrintTo(w, avc)
			_, _ = fmt.Fprintf(w, "\n")
			return nil
		})
	}
}

func onScript(g *gocui.Gui, t *flv.ScriptTag, w io.Writer) {
	decoder := amf.NewDecoder(amf.Version0)
	buf := bytes.NewBuffer(t.Bytes)
	got, err := decoder.DecodeBatch(buf)
	if err != nil {
		showError(g, "Parse metadata fail, err %v\n", err)
		return
	}
	if w != nil {
		submitEvent(func(gui *gocui.Gui) error {
			_, _ = fmt.Fprintf(w, color.RedString("Script Tag:\n"))
			prettyPrintTo(w, got)
			_, _ = fmt.Fprintf(w, color.RedString("\nData:\n"))
			prettyPrintTo(w, t.Data())
			return nil
		})
	} else {
		submitEvent(func(gui *gocui.Gui) error {
			timestampView, _ := gui.View(TimestampViewName)
			_, _ = fmt.Fprintf(timestampView, "{SCRIPT} %7d %7d %7d %7d\n", t.StreamID, t.PTS, t.PTS, len(t.Data()))
			latestTimestampView, _ := gui.View(LatestTimestampViewName)
			_, _ = fmt.Fprintf(latestTimestampView, "{SCRIPT} %7d %7d %7d %7d\n", t.StreamID, t.PTS, t.PTS, len(t.Data()))
			return nil
		})
	}
}

func onAudio(g *gocui.Gui, t *flv.AudioTag, w io.Writer) {
	label := "{ AUDIO}"
	if t.SoundFormat == flv.AAC && t.PacketType == flv.SequenceHeader {
		label = "{   AAC}"
		onAAC(g, t, w)
	}
	if w != nil {
		submitEvent(func(gui *gocui.Gui) error {
			_, _ = fmt.Fprintf(w, color.RedString("Audio Tag:\n"))
			prettyPrintTo(w, t.Data())
			return nil
		})
	} else {
		submitEvent(func(gui *gocui.Gui) error {
			timestampView, _ := gui.View(TimestampViewName)
			_, _ = fmt.Fprintf(timestampView, "%s %7d %7d %7d %7d %s %s %s %s\n",
				label, t.StreamID, t.PTS, t.PTS, len(t.Data()), t.SoundFormat.String(), t.Channels.String(), t.BitPerSample.String(), t.SampleRate.String())
			latestTimestampView, _ := gui.View(LatestTimestampViewName)
			_, _ = fmt.Fprintf(latestTimestampView, "%s %7d %7d %7d %7d %s %s %s %s\n",
				label, t.StreamID, t.PTS, t.PTS, len(t.Data()), t.SoundFormat.String(), t.Channels.String(), t.BitPerSample.String(), t.SampleRate.String())
			return nil
		})
	}
}

func onVideo(g *gocui.Gui, t *flv.VideoTag, w io.Writer) {
	label := "{ VIDEO}"
	if t.PacketType == flv.SequenceHeader {
		label = "{   AVC}"
		if t.CodecID == flv.H265 {
			label = "{  HAVC}"
		}
		onAVC(g, t, w)
	}
	if w != nil {
		submitEvent(func(gui *gocui.Gui) error {
			_, _ = fmt.Fprintf(w, color.RedString("Video Tag:\n"))
			prettyPrintTo(w, t.Data())
			return nil
		})
	} else {
		submitEvent(func(gui *gocui.Gui) error {
			timestampView, _ := gui.View(TimestampViewName)

			_, _ = fmt.Fprintf(timestampView, "%s %7d %7d %7d %7d %s %s\n",
				label, t.StreamID, t.PTS, t.DTS, len(t.Data()), t.FrameType.String(), t.CodecID.String())
			latestTimestampView, _ := gui.View(LatestTimestampViewName)
			_, _ = fmt.Fprintf(latestTimestampView, "%s %7d %7d %7d %7d %s %s\n",
				label, t.StreamID, t.PTS, t.DTS, len(t.Data()), t.FrameType.String(), t.CodecID.String())
			return nil
		})
	}
}
