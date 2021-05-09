package codec

import "errors"

// AACAudioSpecificConfig
//
// 5 bits: object type
// if (object type == 31)
//    6 bits + 32: object type
// 4 bits: frequency index
// if (frequency index == 15)
//    24 bits: frequency
// 4 bits: channel configuration
// var bits: AOT Specific Config
type AACAudioSpecificConfig struct {
	ObjectType byte
	SampleRate byte
	Channel    byte
}

func (aac *AACAudioSpecificConfig) Read(data []byte) error {
	if len(data) < 2 {
		return errors.New("aac audio specific config data invalid")
	}
	aac.ObjectType = data[0] >> 3
	aac.SampleRate = (data[0] & 0x07 << 1) | (data[1] >> 7)
	aac.Channel = (data[1] >> 3) & 0x0f
	return nil
}
