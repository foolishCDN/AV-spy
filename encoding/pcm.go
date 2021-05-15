package encoding

import (
	"encoding/binary"
	"errors"
)

const (
	ChannelMono = iota + 1
	ChannelStereo
)

type SampleSlice interface {
	Type() int
}

type Mono struct {
	Samples []float64
}

func (m *Mono) Type() int {
	return ChannelMono
}

type Stereo struct {
	LSamples []float64
	RSamples []float64
}

func (s *Stereo) Type() int {
	return ChannelStereo
}

type PCM struct {
	NumOfChannels int
	SampleRate    int // 8000, 44100, 48000, etc
	BitPerSample  int
}

func (pcm *PCM) GetSampleParser() (func([]byte) SampleSlice, error) {
	parser, err := pcm.parser()
	if err != nil {
		return nil, err
	}
	bytePerSample := pcm.BitPerSample / 8
	if pcm.NumOfChannels == 1 {
		return func(data []byte) SampleSlice {
			total := len(data) / bytePerSample
			sample := make([]float64, 0)
			for i := 0; i < total; i += bytePerSample {
				sample = append(sample, parser(data))
			}
			return &Mono{
				Samples: sample,
			}
		}, nil
	}
	if pcm.NumOfChannels == 2 {
		return func(data []byte) SampleSlice {
			total := len(data) / bytePerSample
			lSample := make([]float64, 0)
			rSample := make([]float64, 0)
			for i := 0; i < total; i += bytePerSample * 2 {
				l := parser(data[i : i+bytePerSample])
				r := parser(data[i+bytePerSample : i+bytePerSample*2])
				lSample = append(lSample, l)
				rSample = append(rSample, r)
			}
			return &Stereo{
				LSamples: lSample,
				RSamples: rSample,
			}

		}, nil
	}
	return nil, errors.New("parser: unsupported channel")
}

func (pcm *PCM) parser() (func([]byte) float64, error) {
	switch pcm.BitPerSample {
	case 8:
		return int8Parser, nil
	case 16:
		return int16Parser, nil
	case 32:
		return int32Parser, nil
	}
	return nil, errors.New("parser: unsupported pcm")
}

func (pcm *PCM) getBound() (upper, lower float64, err error) {
	switch pcm.BitPerSample {
	case 8:
		upper = 255
		lower = 0
	case 16:
		upper = 32767
		lower = -32768
	case 32:
		upper = 2147483647
		lower = -2147483648
	default:
		err = errors.New("bound: unsupported bit per sample")
	}
	return upper, lower, err
}

func int8Parser(data []byte) float64 {
	return float64(data[0])
}

func int16Parser(data []byte) float64 {
	return float64(binary.LittleEndian.Uint16(data))
}

func int32Parser(data []byte) float64 {
	return float64(binary.LittleEndian.Uint32(data))
}
