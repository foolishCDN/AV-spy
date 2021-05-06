package codec

import (
	"bytes"
	"encoding/binary"
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
	LenOfSPS int
	SPS      []byte

	NumOfPPS byte
	LenOfPPS int
	PPS      []byte
}

func (avc *AVCDecoderConfigurationRecord) Read(data []byte) error {
	if len(data) < 9 {
		return errors.New("sps data is nil")
	}
	avc.ConfigurationVersion = data[0]
	avc.AVCProfileIndication = data[1]
	avc.ProfileCompatibility = data[2]
	avc.AVCLevelIndication = data[3]
	avc.LengthSizeMinusOne = data[4] & 0x03

	avc.NumOfSPS = data[5] & 0x1f
	avc.LenOfSPS = int(data[6])<<8 | int(data[7])
	if len(data[8:]) < avc.LenOfSPS || avc.LenOfSPS <= 0 {
		return errors.New("sps error")
	}
	avc.SPS = data[8 : 8+avc.LenOfSPS]

	ppsData := data[8+avc.LenOfSPS:]
	if len(ppsData) <= 3 {
		return errors.New("pps data is nil")
	}
	avc.NumOfPPS = ppsData[0]
	avc.LenOfPPS = int(ppsData[1])<<8 | int(ppsData[2])
	if len(ppsData[3:]) < avc.LenOfPPS || avc.LenOfPPS <= 0 {
		return errors.New("pps error")
	}
	avc.PPS = ppsData[3 : avc.LenOfPPS+3]
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

	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(avc.LenOfSPS))
	buf.Write(b)
	buf.Write(avc.SPS)

	buf.WriteByte(avc.NumOfPPS)
	binary.BigEndian.PutUint16(b, uint16(avc.LenOfPPS))
	buf.Write(b)
	buf.Write(avc.PPS)
	return buf.Bytes()
}
