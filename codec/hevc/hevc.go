package hevc

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type HEVCNALU struct {
	ArrayCompleteness byte
	NALUnitType       byte
	NALUs             [][]byte
}

// HEVCDecoderConfigurationRecord
// ISO/IEC-14496-15 8.3.3.1
//
//	aligned(8) class HEVCDecoderConfigurationRecord {
//		unsigned int(8) configurationVersion = 1;
//		unsigned int(2) general_profile_space;
//		unsigned int(1) general_tier_flag;
//		unsigned int(5) general_profile_idc;
//		unsigned int(32) general_profile_compatibility_flags;
//		unsigned int(48) general_constraint_indicator_flags;
//		unsigned int(8) general_level_idc;
//		bit(4) reserved = '1111'b;
//		unsigned int(12) min_spatial_segmentation_idc;
//		bit(6) reserved = '111111'b;
//		unsigned int(2) parallelismType;
//		bit(6) reserved = '111111'b;
//		unsigned int(2) chromaFormat;
//		bit(5) reserved = '11111'b;
//		unsigned int(3) bitDepthLumaMinus8;
//		bit(5) reserved = '11111'b;
//		unsigned int(3) bitDepthChromaMinus8;
//		bit(16) avgFrameRate;
//		bit(2) constantFrameRate;
//		bit(3) numTemporalLayers;
//		bit(1) temporalIdNested;
//		unsigned int(2) lengthSizeMinusOne;
//		unsigned int(8) numOfArrays;
//		for (j=0; j < numOfArrays; j++) {
//			bit(1) array_completeness;
//			unsigned int(1) reserved = 0;
//			unsigned int(6) NAL_unit_type;
//			unsigned int(16) numNalus;
//			for (i=0; i< numNalus; i++) {
//				unsigned int(16) nalUnitLength;
//				bit(8*nalUnitLength) nalUnit;
//			}
//		}
//	}
type HEVCDecoderConfigurationRecord struct {
	ConfigurationVersion             byte
	GeneralProfileSpace              byte
	GeneralTierFlag                  byte
	GeneralProfileIDC                byte
	GeneralProfileCompatibilityFlags uint32
	GeneralConstraintIndicatorFlags  uint64
	GeneralLevelIdc                  byte
	MinSpatialSegmentationIdc        uint16
	ParallelismType                  byte
	ChromaFormat                     byte
	BitDepthLumaMinus8               byte
	BitDepthChromaMinus8             byte
	AvgFrameRate                     uint16
	ConstantFrameRate                byte
	NumTemporalLayers                byte
	TemporalIdNested                 byte
	LengthSizeMinusOne               byte

	NALUs []HEVCNALU
}

func (hevc *HEVCDecoderConfigurationRecord) Read(data []byte) error {
	total := len(data)
	if total < 23 {
		return fmt.Errorf("invalid HEVCDecoderConfigurationRecord: %v", hex.Dump(data))
	}
	hevc.ConfigurationVersion = data[0]
	hevc.GeneralProfileSpace = (data[1] >> 6) & 0x03
	hevc.GeneralTierFlag = (data[1] >> 5) & 0x01
	hevc.GeneralProfileIDC = (data[1]) & 0x01f
	hevc.GeneralProfileCompatibilityFlags = binary.BigEndian.Uint32(data[2:])
	hevc.GeneralConstraintIndicatorFlags = uint64(data[6])<<40 | uint64(data[7])<<32 | uint64(data[8])<<24 | uint64(data[9])<<16 | uint64(data[10])<<8 | uint64(data[11])
	hevc.GeneralLevelIdc = data[12]
	hevc.MinSpatialSegmentationIdc = uint16(data[13])&0x0f<<8 | uint16(data[14])
	hevc.ParallelismType = data[15] & 0x03
	hevc.ChromaFormat = data[16] & 0x03
	hevc.BitDepthLumaMinus8 = data[17] & 0x07
	hevc.BitDepthChromaMinus8 = data[18] & 0x07
	hevc.AvgFrameRate = binary.BigEndian.Uint16(data[19:])
	hevc.ConstantFrameRate = (data[21] >> 6) & 0x03
	hevc.NumTemporalLayers = (data[21] >> 3) & 0x07
	hevc.TemporalIdNested = (data[21] >> 2) & 0x01
	hevc.LengthSizeMinusOne = data[21] & 0x03

	offset := 23
	for i := byte(0); i < data[22]; i++ {
		if offset+3 > total {
			return fmt.Errorf("invalid array: %v", hex.Dump(data))
		}
		nalu := HEVCNALU{}
		nalu.ArrayCompleteness = (data[offset] >> 7) & 0x01
		nalu.NALUnitType = data[offset] & 0x3f

		num := binary.BigEndian.Uint16(data[offset+1:])
		offset += 3
		for j := uint16(0); j < num; j++ {
			if offset+2 > total {
				return fmt.Errorf("invalid nalu: %v", hex.Dump(data))
			}
			k := int(binary.BigEndian.Uint16(data[offset:]))
			if offset+2+k > total {
				return fmt.Errorf("invalid nalu length: %v", hex.Dump(data))
			}
			nalu.NALUs = append(nalu.NALUs, data[offset+2:offset+2+k])
			offset += 2 + k
		}
		hevc.NALUs = append(hevc.NALUs, nalu)
	}

	return nil
}

func (hevc *HEVCDecoderConfigurationRecord) Write() []byte {
	buf := make([]byte, 23)
	buf[0] = hevc.ConfigurationVersion
	buf[1] = (hevc.GeneralProfileSpace&0x03)<<6 | (hevc.GeneralTierFlag&0x01)<<5 | (hevc.GeneralProfileIDC)&0x0F
	binary.BigEndian.PutUint64(buf[4:], hevc.GeneralConstraintIndicatorFlags)
	binary.BigEndian.PutUint32(buf[2:], hevc.GeneralProfileCompatibilityFlags)
	buf[12] = hevc.GeneralLevelIdc
	binary.BigEndian.PutUint16(buf[13:], 0xF000|hevc.MinSpatialSegmentationIdc)
	buf[15] = 0xFC | hevc.ParallelismType
	buf[16] = 0xFC | hevc.ChromaFormat
	buf[17] = 0xF8 | hevc.BitDepthLumaMinus8
	buf[18] = 0xF8 | hevc.BitDepthChromaMinus8
	binary.BigEndian.PutUint16(buf[19:], hevc.AvgFrameRate)
	buf[21] = hevc.ConstantFrameRate<<6 | (hevc.NumTemporalLayers&0x07)<<3 | (hevc.TemporalIdNested&0x01)<<2 | hevc.LengthSizeMinusOne&0x03
	buf[22] = byte(len(hevc.NALUs))

	for _, nalu := range hevc.NALUs {
		n := len(nalu.NALUs)
		buf = append(buf, (nalu.ArrayCompleteness<<7)|(nalu.NALUnitType&0x3F))
		buf = append(buf, byte(n>>8), byte(n))
		for _, item := range nalu.NALUs {
			l := len(item)
			buf = append(buf, byte(l>>8), byte(l))
			buf = append(buf, item...)
		}
	}
	return buf
}
