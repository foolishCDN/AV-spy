package hevc

import "github.com/foolishCDN/AV-spy/utils"

type SubLayerHrdParameters struct {
	BitRateValueMinus1   []uint
	CpbSizeValueMinus1   []uint
	CpbSizeDuValueMinus1 []uint
	BitRateDuValueMinus1 []uint
	CbrFlag              []bool
}

type HRD struct {
	NalHrdParametersPresentFlag    bool
	VclHrdParametersPresentFlag    bool
	SubPicHrdParametersPresentFlag bool

	TickDivisorMinus2                      uint8
	DuCpbRemovalDelayIncrementLengthMinus1 uint8
	SubPicCpbParamsInPicTimingSeiFlag      bool
	DpbOutputDelayDuLengthMinus1           uint8

	BitRateScale   uint8
	CpbSizeScale   uint8
	CpbSizeDuScale uint8

	InitialCpbRemovalDelayLengthMinus1 uint8
	AuCpbRemovalDelayLengthMinus1      uint8
	DpbOutputDelayLengthMinus1         uint8

	FixedPicRateGeneralFlag     []bool
	FixedPicRateWithinCvsFlag   []bool
	ElementalDurationInTcMinus1 []uint
	LowDelayHrdFlag             []bool
	CpbCntMinus1                []uint

	SubLayerHrdParameters []*SubLayerHrdParameters
}

func ParseHRD(reader *utils.BitReader, commonInfPresentFlag bool, maxNumSubLayerMinus1 uint8) *HRD {
	hrd := new(HRD)
	if commonInfPresentFlag {
		hrd.NalHrdParametersPresentFlag = reader.ReadFlag()
		hrd.VclHrdParametersPresentFlag = reader.ReadFlag()

		if hrd.NalHrdParametersPresentFlag || hrd.VclHrdParametersPresentFlag {
			hrd.SubPicHrdParametersPresentFlag = reader.ReadFlag()
			if hrd.SubPicHrdParametersPresentFlag {
				hrd.TickDivisorMinus2 = reader.ReadBitsUint8(8)
				hrd.DuCpbRemovalDelayIncrementLengthMinus1 = reader.ReadBitsUint8(5)
				hrd.SubPicCpbParamsInPicTimingSeiFlag = reader.ReadFlag()
				hrd.DpbOutputDelayDuLengthMinus1 = reader.ReadBitsUint8(5)
			}

			hrd.BitRateScale = reader.ReadBitsUint8(4)
			hrd.CpbSizeScale = reader.ReadBitsUint8(4)

			if hrd.SubPicHrdParametersPresentFlag {
				hrd.CpbSizeDuScale = reader.ReadBitsUint8(4)
			}
			hrd.InitialCpbRemovalDelayLengthMinus1 = reader.ReadBitsUint8(5)
			hrd.AuCpbRemovalDelayLengthMinus1 = reader.ReadBitsUint8(5)
			hrd.DpbOutputDelayLengthMinus1 = reader.ReadBitsUint8(5)
		}
	}

	for i := 0; i < int(maxNumSubLayerMinus1); i++ {
		hrd.FixedPicRateGeneralFlag = append(hrd.FixedPicRateGeneralFlag, reader.ReadFlag())
		if !hrd.FixedPicRateGeneralFlag[i] {
			hrd.FixedPicRateWithinCvsFlag = append(hrd.FixedPicRateWithinCvsFlag, reader.ReadFlag())
		} else {
			hrd.FixedPicRateWithinCvsFlag = append(hrd.FixedPicRateWithinCvsFlag, true)
		}

		if hrd.FixedPicRateWithinCvsFlag[i] {
			hrd.ElementalDurationInTcMinus1 = append(hrd.ElementalDurationInTcMinus1, reader.ReadUE())
		} else {
			hrd.LowDelayHrdFlag = append(hrd.LowDelayHrdFlag, reader.ReadFlag())
		}

		if !hrd.LowDelayHrdFlag[i] {
			hrd.CpbCntMinus1 = append(hrd.CpbCntMinus1, reader.ReadUE())
		}

		if hrd.NalHrdParametersPresentFlag {
			hrd.SubLayerHrdParameters = append(hrd.SubLayerHrdParameters, hrd.parseSubLayerHrdParameters(reader, int(hrd.CpbCntMinus1[i]), hrd.SubPicHrdParametersPresentFlag))
		}

		if hrd.VclHrdParametersPresentFlag {
			hrd.SubLayerHrdParameters = append(hrd.SubLayerHrdParameters, hrd.parseSubLayerHrdParameters(reader, int(hrd.CpbCntMinus1[i]), hrd.SubPicHrdParametersPresentFlag))
		}
	}

	return hrd
}

func (hrd *HRD) parseSubLayerHrdParameters(reader *utils.BitReader, cpbCntMinus1 int, subPicHrdParamsPresentFlag bool) *SubLayerHrdParameters {
	subHrd := new(SubLayerHrdParameters)
	for i := 0; i <= cpbCntMinus1; i++ {
		subHrd.BitRateValueMinus1 = append(subHrd.BitRateValueMinus1, reader.ReadUE())
		subHrd.CpbSizeValueMinus1 = append(subHrd.CpbSizeValueMinus1, reader.ReadUE())
		if subPicHrdParamsPresentFlag {
			subHrd.BitRateDuValueMinus1 = append(subHrd.BitRateDuValueMinus1, reader.ReadUE())
			subHrd.CpbSizeDuValueMinus1 = append(subHrd.CpbSizeDuValueMinus1, reader.ReadUE())
		}
		subHrd.CbrFlag = append(subHrd.CbrFlag, reader.ReadFlag())
	}
	return subHrd
}
