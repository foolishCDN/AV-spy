package hevc

import (
	"github.com/foolishCDN/AV-spy/codec/avc"
	"github.com/foolishCDN/AV-spy/utils"
)

const (
	NalVPS       = 32
	NalSPS       = 33
	NalPPS       = 34
	NalAUD       = 35
	NalSEIPrefix = 39
	NalSEISuffix = 40
)

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

type NALUType int

const (
	NALUTypeInvalid NALUType = iota // Invalid
	NALUTypeHVCC                    // HVCC
	NALUTypeAnnexB                  // AnnexB
)

func SplitNALUS(b []byte) ([][]byte, NALUType) {
	nalus, t := avc.SplitNALUs(b)
	switch t {
	case avc.NALUTypeInvalid:
		return nil, NALUTypeInvalid
	case avc.NALUTypeAVCC:
		return nalus, NALUTypeHVCC
	case avc.NALUTypeAnnexB:
		return nalus, NALUTypeAnnexB
	}
	return nil, NALUTypeInvalid
}
