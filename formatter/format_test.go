package formatter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTemplate(t *testing.T) {
	got := NewTemplate("$stream_type $stream_id:%7d $pts:%7d $dts:%7d $size:%7d $frame_type $codec_id").Format(
		map[ElementName]interface{}{
			ElementStreamType:     "video",
			ElementStreamID:       2,
			ElementPTS:            1010,
			ElementDTS:            1000,
			ElementSize:           4531,
			ElementVideoFrameType: "KeyFrame",
			ElementVideoCodecID:   "H265",
		})
	assert.Equal(t, "video       2    1010    1000    4531 KeyFrame H265", got)

	got = NewTemplate("$stream_type $stream_id:%7d $pts:%7d $dts:%7d $size:%7d $sound_format $channels $sound_size $sample_rate").Format(
		map[ElementName]interface{}{
			ElementStreamType:        "audio",
			ElementStreamID:          2,
			ElementPTS:               1000,
			ElementDTS:               1000,
			ElementSize:              123,
			ElementAudioSoundFormant: "MP3",
			ElementAudioChannels:     "stereo",
			ElementAudioSampleRate:   "44KHz",
			ElementAudioSoundSize:    "16",
		})
	assert.Equal(t, "audio       2    1000    1000     123 MP3 stereo 16 44KHz", got)
}
