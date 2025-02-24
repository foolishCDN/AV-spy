package avc

import "github.com/foolishCDN/AV-spy/utils"

type NALUHeader struct {
	ForbiddenZeroBit bool
	NalRefIdc        uint8
	NalUnitType      uint8
}

func ParseNALUHeader(reader *utils.BitReader) *NALUHeader {
	nalu := new(NALUHeader)
	nalu.ForbiddenZeroBit = reader.ReadFlag()
	nalu.NalRefIdc = reader.ReadBitsUint8(2)
	nalu.NalUnitType = reader.ReadBitsUint8(5)
	return nalu
}
