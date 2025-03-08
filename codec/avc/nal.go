package avc

import (
	"bytes"
	"encoding/binary"

	"github.com/foolishCDN/AV-spy/utils"
)

const (
	NalSEI = 6
)

type NALUType int

const (
	NALUTypeInvalid NALUType = iota // Invalid
	NALUTypeAVCC                    // AVCC
	NALUTypeAnnexB                  // AnnexB
)

var (
	startCode = []byte{0x00, 0x00, 0x01}
)

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

func SplitNALUs(b []byte) ([][]byte, NALUType) {
	if len(b) < 4 {
		return nil, NALUTypeInvalid
	}
	nalus := SplitNALUsAVCC(b)
	if len(nalus) > 0 {
		return nalus, NALUTypeAVCC
	}

	nalus = SplitNALUsAnnexB(b)
	if len(nalus) > 0 {
		return nalus, NALUTypeAnnexB
	}

	return nil, NALUTypeInvalid
}

func SplitNALUsAVCC(b []byte) [][]byte {
	length := binary.BigEndian.Uint32(b)
	if length > uint32(len(b)) {
		return nil
	}
	data := b[4:]
	nalus := make([][]byte, 0)
	for {
		nalus = append(nalus, data[:length])
		data = data[length:]
		if len(data) < 4 {
			break
		}
		length = binary.BigEndian.Uint32(data)
		data = data[4:]
		if length > uint32(len(data)) {
			break
		}
	}
	if len(data) == 0 {
		return nalus
	}
	return nil
}

func SplitNALUsAnnexB(b []byte) [][]byte {
	val3 := utils.BigEndianUint24(b)
	val4 := binary.BigEndian.Uint32(b)
	// StartCode 0x00_00_01 or 0x00_00_00_01
	if val3 != 1 && val4 != 1 {
		return nil
	}
	var i, j, k int
	nalus := make([][]byte, 0)
	end := len(b)

	for i = bytes.Index(b, startCode); i != -1 && i < end; i = j {
		i += 3 // advance 0x00_00_01

		if i < end {
			j = bytes.Index(b[i:], startCode)
		} else {
			j = -1
		}

		if j == -1 {
			j = end
		} else {
			j += i
		}

		// trim right '0' for first 0x00 in 0x00_00_00_01
		for k = j; k > i && b[k-1] == 0; {
			k--
		}
		if k > i {
			nalus = append(nalus, b[i:k])
		}
	}

	return nalus
}
