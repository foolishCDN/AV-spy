package formatter

import (
	"fmt"
	"sort"
	"strings"
)

type ElementName string

var definedElements []ElementName

var DefaultVideoTemplate Template
var DefaultAudioTemplate Template
var DefaultScriptTemplate Template

func init() {
	definedElements = []ElementName{
		ElementStreamType,
		ElementStreamID,
		ElementDTS,
		ElementPTS,
		ElementSize,

		ElementAudioSoundFormant,
		ElementAudioChannels,
		ElementAudioSoundSize,
		ElementAudioSampleRate,

		ElementVideoFrameType,
		ElementVideoCodecID,
	}
	sort.Slice(definedElements, func(i, j int) bool {
		return len(definedElements[i]) > len(definedElements[j]) // 按字符串长度 降序 排列
	})
}

const (
	ElementStreamType ElementName = "stream_type" // AUDIO/VIDEO/SCRIPT AAC/AVC/HEVC
	ElementStreamID   ElementName = "stream_id"
	ElementPTS        ElementName = "pts"
	ElementDTS        ElementName = "dts"
	ElementSize       ElementName = "size"

	ElementAudioSoundFormant ElementName = "sound_format"
	ElementAudioChannels     ElementName = "channels"
	ElementAudioSoundSize    ElementName = "sound_size"
	ElementAudioSampleRate   ElementName = "sample_rate"

	ElementVideoFrameType ElementName = "frame_type"
	ElementVideoCodecID   ElementName = "codec_id"
)

type Template struct {
	Template    string
	OriTemplate string
	Elements    []Element
}

func (t *Template) Format(vars map[ElementName]interface{}) string {
	values := make([]interface{}, 0, len(t.Elements))
	for _, element := range t.Elements {
		v, ok := vars[element.Name]
		if ok {
			values = append(values, v)
		}
	}
	return fmt.Sprintf(t.Template, values...)
}

type Element struct {
	Name        ElementName
	Format      string
	FormatExist bool
}

func splitElements(template string) []Element {
	elements := strings.Split(template, "$")
	res := make([]Element, 0, len(elements))
	for _, element := range elements {
		if len(element) == 0 {
			continue
		}
		for _, defined := range definedElements {
			if strings.Contains(element, string(defined)) {
				var format string
				index := strings.Index(element, ":")
				if index != -1 {
					format = element[index+1:]
				}
				res = append(res, Element{
					Name:   defined,
					Format: format,
				})
				break
			}
		}
	}
	return res
}

func NewTemplate(origin string) *Template {
	template := origin
	elements := splitElements(template)
	for _, element := range elements {
		replaced := "$" + string(element.Name)
		if len(element.Format) > 0 {
			replaced = replaced + ":" + element.Format
		} else {
			element.Format = "%s"
		}
		template = strings.Replace(template, replaced, element.Format, 1)
	}
	return &Template{
		Elements:    elements,
		Template:    template,
		OriTemplate: origin,
	}
}
