package hevc

import "github.com/foolishCDN/AV-spy/utils"

type NALUHeader struct {
	ForbiddenZeroBit   bool
	NALUnitType        uint8
	NUHLayerID         uint8
	NUHTemporalIDPlus1 uint8
}

func ParseNALUHeader(reader *utils.BitReader) *NALUHeader {
	nalu := new(NALUHeader)
	nalu.ForbiddenZeroBit = reader.ReadFlag()
	nalu.NALUnitType = reader.ReadBitsUint8(6)
	nalu.NUHLayerID = reader.ReadBitsUint8(6)
	nalu.NUHTemporalIDPlus1 = reader.ReadBitsUint8(3)
	return nalu
}
