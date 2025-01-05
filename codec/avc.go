package codec

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

var profiles = []byte{100, 110, 122, 144}

// AVCDecoderConfigurationRecord
// ISO/IEC-14496-15 5.2.4.1
// AVCDecoderConfigurationRecord Definition
//
//	aligned(8) class AVCDecoderConfigurationRecord {
//		unsigned int(8) configurationVersion = 1;
//		unsigned int(8) AVCProfileIndication;
//		unsigned int(8) profile_compatibility;
//		unsigned int(8) AVCLevelIndication;
//		bit(6) reserved = '111111'b;
//		unsigned int(2) lengthSizeMinusOne;
//		bit(3) reserved = '111'b;
//
//		unsigned int(5) numOfSequenceParameterSets;
//		for (i=0; i< numOfSequenceParameterSets; i++) {
//			unsigned int(16) sequenceParameterSetLength ;
//			bit(8*sequenceParameterSetLength) sequenceParameterSetNALUnit;
//		}
//
//		unsigned int(8) numOfPictureParameterSets;
//		for (i=0; i< numOfPictureParameterSets; i++) {
//			unsigned int(16) pictureParameterSetLength;
//			bit(8*pictureParameterSetLength) pictureParameterSetNALUnit;
//		}
//
//		if (profile_idc == 100 || profile_idc == 110 || profile_idc == 122 || profile_idc == 144) {
//			bit(6) reserved = '111111'b
//			unsigned int(2) chroma_format;
//			bit(5) reserved = '11111'b;
//			unsigned int(3) bit_depth_luma_minus8;
//			bit(5) reserved = '11111'b;
//			unsigned int(3) bit_depth_chroma_minus8;
//			unsigned int(8) num_of_sequence_parameter_set_ext;
//			for (i=0; i< num_of_sequence_parameter_set_ext; i++) {
//				unsigned int(16) sequenceParameterSetExtLength;
//				bit(8*sequenceParameterSetExtLength) sequenceParameterSetExtNALUnit;
//			}
//		}
//	}
type AVCDecoderConfigurationRecord struct {
	ConfigurationVersion byte
	AVCProfileIndication byte
	ProfileCompatibility byte
	AVCLevelIndication   byte
	LengthSizeMinusOne   byte
	SPS                  [][]byte
	PPS                  [][]byte

	ChromaFormat         byte
	BitDepthLumaMinus8   byte
	BitDepthChromaMinus8 byte

	NumOfSPSExt byte
	SPSExt      [][]byte
}

func (avc *AVCDecoderConfigurationRecord) Read(data []byte) error {
	total := len(data)
	if total < 6 {
		return fmt.Errorf("invalid AVCDecoderConfigurationRecord: %v", hex.Dump(data))
	}
	avc.ConfigurationVersion = data[0]
	avc.AVCProfileIndication = data[1]
	avc.ProfileCompatibility = data[2]
	avc.AVCLevelIndication = data[3]
	avc.LengthSizeMinusOne = data[4] & 0x03

	numOfSPS := data[5] & 0x1f

	offset := 6
	for i := byte(0); i < numOfSPS && offset+2 < total; i++ {
		length := int(data[offset])<<8 | int(data[offset+1])
		if offset+2+length > total {
			return fmt.Errorf("invalid sps: %v", hex.Dump(data))
		}
		avc.SPS = append(avc.SPS, data[offset+2:offset+2+length])
		offset += 2 + length
	}
	if offset > total {
		return fmt.Errorf("no pps: %v", hex.Dump(data))
	}

	if len(data[offset:]) == 0 {
		return fmt.Errorf("empty pps: %v", hex.Dump(data))
	}
	numOfPPS := data[offset]
	offset++
	for i := byte(0); i < numOfPPS && offset+2 < total; i++ {
		length := int(data[offset])<<8 | int(data[offset+1])
		if offset+2+length > total {
			return fmt.Errorf("invalid pps: %v", hex.Dump(data))
		}
		avc.PPS = append(avc.PPS, data[offset+2:offset+2+length])
		offset += 2 + length
	}

	if bytes.IndexByte(profiles, avc.AVCProfileIndication) == -1 {
		return nil
	}
	if offset+4 > total {
		return nil
	}
	data = data[offset:]
	avc.ChromaFormat = data[0] & 0x02
	avc.BitDepthLumaMinus8 = data[1] & 0x03
	avc.BitDepthChromaMinus8 = data[2] & 0x03
	avc.NumOfSPSExt = data[3]

	for i := byte(0); i < avc.NumOfSPSExt && offset+2 < total; i++ {
		length := int(data[offset])<<8 | int(data[offset+1])
		if offset+2+length > total {
			return fmt.Errorf("invalid sps ext: %v", hex.Dump(data))
		}
		avc.SPSExt = append(avc.SPSExt, data[offset+2:offset+2+length])
		offset += 2 + length
	}
	return nil
}

func (avc *AVCDecoderConfigurationRecord) Write() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 9))
	buf.WriteByte(avc.ConfigurationVersion)
	buf.WriteByte(avc.AVCProfileIndication)
	buf.WriteByte(avc.ProfileCompatibility)
	buf.WriteByte(avc.AVCLevelIndication)
	buf.WriteByte(avc.LengthSizeMinusOne | 0xfc)

	buf.WriteByte(byte(len(avc.SPS) | 0xe0))
	for i := range avc.SPS {
		n := len(avc.SPS[i])
		buf.WriteByte(byte(n >> 8))
		buf.WriteByte(byte(n))
		buf.Write(avc.SPS[i])
	}

	buf.WriteByte(byte(len(avc.PPS)))
	for i := range avc.PPS {
		n := len(avc.PPS[i])
		buf.WriteByte(byte(n >> 8))
		buf.WriteByte(byte(n))
		buf.Write(avc.PPS[i])
	}

	if bytes.IndexByte(profiles, avc.AVCProfileIndication) != -1 {
		buf.WriteByte(avc.ChromaFormat | 0xFC)
		buf.WriteByte(avc.BitDepthLumaMinus8 | 0xF8)
		buf.WriteByte(avc.BitDepthChromaMinus8 | 0xF8)
		buf.WriteByte(avc.NumOfSPSExt)

		for i := range avc.SPSExt {
			n := len(avc.SPSExt[i])
			buf.WriteByte(byte(n >> 8))
			buf.WriteByte(byte(n))
			buf.Write(avc.SPSExt[i])
		}
	}

	return buf.Bytes()
}
