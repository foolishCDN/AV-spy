package codec

import (
	"bytes"
	"errors"
)

// AVCDecoderConfigurationRecord Definition
//
//aligned(8) class AVCDecoderConfigurationRecord {
//	unsigned int(8) configurationVersion = 1;
//	unsigned int(8) AVCProfileIndication;
//	unsigned int(8) profile_compatibility;
//	unsigned int(8) AVCLevelIndication;
//	bit(6) reserved = ‘111111’b;
//	unsigned int(2) lengthSizeMinusOne;
//	bit(3) reserved = ‘111’b;
//
//	unsigned int(5) numOfSequenceParameterSets;
//	for (i=0; i< numOfSequenceParameterSets; i++) {
//		unsigned int(16) sequenceParameterSetLength ;
//		bit(8*sequenceParameterSetLength) sequenceParameterSetNALUnit;
//	}
//
//	unsigned int(8) numOfPictureParameterSets;
//	for (i=0; i< numOfPictureParameterSets; i++) {
//		unsigned int(16) pictureParameterSetLength;
//		bit(8*pictureParameterSetLength) pictureParameterSetNALUnit;
//	}
//}

type AVCDecoderConfigurationRecord struct {
	ConfigurationVersion byte
	AVCProfileIndication byte
	ProfileCompatibility byte
	AVCLevelIndication   byte

	LengthSizeMinusOne byte

	NumOfSPS byte
	SPS      [][]byte

	NumOfPPS byte
	PPS      [][]byte
}

func (avc *AVCDecoderConfigurationRecord) Read(data []byte) error {
	total := len(data)
	if total < 6 {
		return errors.New("invalid sps data")
	}
	avc.ConfigurationVersion = data[0]
	avc.AVCProfileIndication = data[1]
	avc.ProfileCompatibility = data[2]
	avc.AVCLevelIndication = data[3]
	avc.LengthSizeMinusOne = data[4] & 0x03

	avc.NumOfSPS = data[5] & 0x1f

	now := 6
	for i := byte(0); i < avc.NumOfSPS && now+2 < total; i++ {
		length := int(data[now])<<8 | int(data[now+1])
		if now+2+length > total {
			return errors.New("invalid sps")
		}
		avc.SPS = append(avc.SPS, data[now+2:now+2+length])
		now += 2 + length
	}
	if now > total {
		return errors.New("no pps")
	}

	if len(data[now:]) == 0 {
		return errors.New("pps data is nil")
	}
	avc.NumOfPPS = data[now]
	now++
	for i := byte(0); i < avc.NumOfPPS && now+2 < total; i++ {
		length := int(data[now])<<8 | int(data[now+1])
		if now+2+length > total {
			return errors.New("invalid pps")
		}
		avc.PPS = append(avc.PPS, data[now+2:now+2+length])
		now += 2 + length
	}
	return nil
}

func (avc *AVCDecoderConfigurationRecord) Write() []byte {
	buf := bytes.NewBuffer(make([]byte, 9))
	buf.WriteByte(avc.ConfigurationVersion)
	buf.WriteByte(avc.AVCProfileIndication)
	buf.WriteByte(avc.ProfileCompatibility)
	buf.WriteByte(avc.AVCLevelIndication)
	buf.WriteByte(avc.LengthSizeMinusOne | 0xfc)

	buf.WriteByte(avc.NumOfSPS | 0xe0)
	for i := range avc.SPS {
		n := len(avc.SPS[i])
		buf.WriteByte(byte(n >> 8))
		buf.WriteByte(byte(n))
		buf.Write(avc.SPS[i])
	}

	buf.WriteByte(avc.NumOfPPS)
	for i := range avc.PPS {
		n := len(avc.PPS[i])
		buf.WriteByte(byte(n >> 8))
		buf.WriteByte(byte(n))
		buf.Write(avc.PPS[i])
	}
	return buf.Bytes()
}
