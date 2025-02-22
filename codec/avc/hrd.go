package avc

import "github.com/foolishCDN/AV-spy/utils"

type HRD struct {
	CpbCntMinus1       uint
	BitRateScale       uint8
	CpbSizeScale       uint8
	BitRateValueMinus1 []uint
	CpbSizeValueMinus1 []uint
	CbrFlag            []bool

	InitialCpbRemovalDelayLengthMinus1 uint8
	CpbRemovalDelayLengthMinus1        uint8
	DpbOutputDelayLengthMinus1         uint8
	TimeOffsetLength                   uint8
}

func ParseHRD(reader *utils.BitReader) HRD {
	hrd := HRD{}

	hrd.CpbCntMinus1 = reader.ReadUE()
	hrd.BitRateScale = reader.ReadBitsUint8(4)
	hrd.CpbSizeScale = reader.ReadBitsUint8(4)

	hrd.BitRateValueMinus1 = make([]uint, hrd.CpbCntMinus1+1)
	hrd.CpbSizeValueMinus1 = make([]uint, hrd.CpbCntMinus1+1)
	hrd.CbrFlag = make([]bool, hrd.CpbCntMinus1+1)
	for i := 0; i <= int(hrd.CpbCntMinus1); i++ {
		hrd.BitRateValueMinus1[i] = reader.ReadUE()
		hrd.CpbSizeValueMinus1[i] = reader.ReadUE()
		hrd.CbrFlag[i] = reader.ReadFlag()
	}
	hrd.InitialCpbRemovalDelayLengthMinus1 = reader.ReadBitsUint8(5)
	hrd.CpbRemovalDelayLengthMinus1 = reader.ReadBitsUint8(5)
	hrd.DpbOutputDelayLengthMinus1 = reader.ReadBitsUint8(5)
	hrd.TimeOffsetLength = reader.ReadBitsUint8(5)
	return hrd
}
