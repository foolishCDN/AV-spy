package hevc

import (
	"encoding/hex"
	"errors"

	"github.com/foolishCDN/AV-spy/utils"
	"github.com/sikasjc/pretty"
	"github.com/sirupsen/logrus"
)

var SubWidthC = [4]int{1 /*4:0:0*/, 2 /*4:2:0*/, 2 /*4:2:2*/, 1 /*4:4:4*/}
var SubHeightC = [4]int{1 /*4:0:0*/, 2 /*4:2:0*/, 1 /*4:2:2*/, 1 /*4:4:4*/}

type SPS struct {
	SPSVideoParameterSetID   uint8
	SPSMaxSubLayersMinus1    uint8
	SPSTemporalIdNestingFlag bool
	ProfileTierLevel         *ProfileTierLevel
	SPSSeqParameterSetId     uint
	ChromaFormatIdc          uint
	SeparateColourPlaneFlag  bool
	PicWidthInLumaSamples    uint
	PicHeightInLumaSamples   uint
	ConformanceWindowFlag    bool
	ConfWinLeftOffset        uint
	ConfWinRightOffset       uint
	ConfWinTopOffset         uint
	ConfWinBottomOffset      uint

	BitDepthLumaMinus8                 uint
	BitDepthChromaMinus8               uint
	Log2MaxPicOrderCntLsbMinus4        uint
	SPSSubLayerOrderingInfoPresentFlag bool

	SPSMaxDecPicBufferingMinus1 []uint
	SPSMaxNumReorderPics        []uint
	SPSMaxLatencyIncreasePlus1  []uint

	Log2MinLumaCodingBlockSizeMinus3     uint
	Log2DiffMaxMinLumaCodingBlockSize    uint
	Log2MinLumaTransformBlockSizeMinus2  uint
	Log2DiffMaxMinLumaTransformBlockSize uint
	MaxTransformHierarchyDepthInter      uint
	MaxTransformHierarchyDepthIntra      uint

	ScalingListEnabledFlag          bool
	SPSScalingListDataPresentFlag   bool
	AmpEnabledFlag                  bool
	SampleAdaptiveOffsetEnabledFlag bool

	PcmEnabledFlag                       bool
	PcmSampleBitDepthLumaMinus1          uint8
	PcmSampleBitDepthChromaMinus1        uint8
	Log2MinPcmLumaCodingBlockSizeMinus3  uint
	Log2DiffMaxMinPcmLumaCodingBlockSize uint
	PcmLoopFilterDisabledFlag            bool

	NumShortTermRefPicSets     uint
	LongTermRefPicsPresentFlag bool
	NumNegativePics            [64]uint
	NumPositivePics            [64]uint
	NumLongTermRefPicsSps      uint

	SPSTemporalMvpEnabledFlag       bool
	StrongIntraSmoothingEnabledFlag bool
	VUIParametersPresentFlag        bool
	VUI                             *VUI
	// TODO: parse more
}

func (sps *SPS) X() int {
	cropUnitX, _ := sps.CropUnit()
	return cropUnitX * int(sps.ConfWinLeftOffset)
}

func (sps *SPS) Y() int {
	_, cropUnitY := sps.CropUnit()
	return cropUnitY * int(sps.ConfWinTopOffset)
}

func (sps *SPS) Width() int {
	cropUnitX, _ := sps.CropUnit()
	return int(sps.PicWidthInLumaSamples) - cropUnitX*int(sps.ConfWinRightOffset) - sps.X()
}

func (sps *SPS) Height() int {
	_, cropUnitY := sps.CropUnit()
	return int(sps.PicHeightInLumaSamples) - cropUnitY*int(sps.ConfWinBottomOffset) - sps.Y()
}

func (sps *SPS) FPS() float64 {
	if sps.VUI == nil {
		return 0
	}
	return sps.VUI.FPS()
}

func (sps *SPS) CropUnit() (int, int) {
	cropUnitX := SubWidthC[sps.ChromaFormatIdc%4]
	cropUnitY := SubHeightC[sps.ChromaFormatIdc%4]
	return cropUnitX, cropUnitY
}

func (sps *SPS) scalingListData(reader *utils.BitReader) {
	for sizeID := 0; sizeID < 4; sizeID++ {
		step := 1
		if sizeID == 3 {
			step = 3
		}
		for matrixID := 0; matrixID < 6; matrixID += step {
			// scaling_list_pred_mode_flag[ sizeId ][ matrixId ]
			scalingListPredModeFlag := reader.ReadFlag()
			if !scalingListPredModeFlag {
				// scaling_list_pred_matrix_id_delta[ sizeId ][ matrixId ]
				reader.ReadUE()
			} else {
				nextCoef := 8
				coefNum := min(64, 1<<(4+(sizeID<<1)))
				if sizeID > 1 {
					// scaling_list_dc_coef_minus8[ sizeId − 2 ][ matrixId ]
					scalingListDcCoefMinus8 := reader.ReadSE()
					// nextCoef = scaling_list_dc_coef_minus8[ sizeId − 2 ][ matrixId ] + 8
					nextCoef = scalingListDcCoefMinus8 + 8
				}

				for i := 0; i < coefNum; i++ {
					scalingListDeltaCoef := reader.ReadSE()
					nextCoef = (nextCoef + scalingListDeltaCoef + 256) % 256
					// ScalingList[ sizeId ][ matrixId ][ i ] = nextCoef
				}
			}
		}
	}
}

func (sps *SPS) stRefPicSet(reader *utils.BitReader, stRpsIdx int) error {
	interRefPicSetPredictionFlag := false
	if stRpsIdx != 0 {
		// inter_ref_pic_set_prediction_flag
		interRefPicSetPredictionFlag = reader.ReadFlag()
	}
	if interRefPicSetPredictionFlag {
		if stRpsIdx == int(sps.NumShortTermRefPicSets) {
			// delta_idx_minus1
			reader.ReadUE()
		}
		// delta_rps_sign
		reader.ReadFlag()
		// abs_delta_rps_minus1
		reader.ReadUE()

		// NumDeltaPocs[ stRpsIdx ] = NumNegativePics[ stRpsIdx ] + NumPositivePics[ stRpsIdx ]
		numDeltaPocs := sps.NumNegativePics[stRpsIdx] + sps.NumPositivePics[stRpsIdx]
		for j := 0; j <= int(numDeltaPocs); j++ {
			// used_by_curr_pic_flag[ j ]
			usedByCurrPicFlag := reader.ReadFlag()
			if !usedByCurrPicFlag {
				// use_delta_flag[ j ]
				reader.ReadFlag()
			}
		}
	} else {
		sps.NumNegativePics[stRpsIdx] = reader.ReadUE()
		sps.NumPositivePics[stRpsIdx] = reader.ReadUE()
		if sps.NumNegativePics[stRpsIdx] > 16 || sps.NumPositivePics[stRpsIdx] > 16 {
			return errors.New("invalid data")
		}
		for i := 0; i < int(sps.NumNegativePics[stRpsIdx]); i++ {
			// delta_poc_s0_minus1[ i ]
			reader.ReadUE()
			// used_by_curr_pic_s0_flag[ i ]
			reader.ReadFlag()
		}
		for i := 0; i < int(sps.NumPositivePics[stRpsIdx]); i++ {
			// delta_poc_s1_minus1[ i ]
			reader.ReadUE()
			// used_by_curr_pic_s1_flag[ i ]
			reader.ReadFlag()
		}
	}
	return nil
}

func ParseSPS(reader *utils.BitReader) (*SPS, error) {
	sps := new(SPS)
	sps.SPSVideoParameterSetID = reader.ReadBitsUint8(4)
	sps.SPSMaxSubLayersMinus1 = reader.ReadBitsUint8(3)
	sps.SPSTemporalIdNestingFlag = reader.ReadFlag()
	sps.ProfileTierLevel = ParseProfileTierLevel(reader, true, sps.SPSMaxSubLayersMinus1)
	sps.SPSSeqParameterSetId = reader.ReadUE()
	sps.ChromaFormatIdc = reader.ReadUE()
	if sps.ChromaFormatIdc == 3 {
		sps.SeparateColourPlaneFlag = reader.ReadFlag()
	}
	sps.PicWidthInLumaSamples = reader.ReadUE()
	sps.PicHeightInLumaSamples = reader.ReadUE()
	sps.ConformanceWindowFlag = reader.ReadFlag()
	if sps.ConformanceWindowFlag {
		sps.ConfWinLeftOffset = reader.ReadUE()
		sps.ConfWinRightOffset = reader.ReadUE()
		sps.ConfWinTopOffset = reader.ReadUE()
		sps.ConfWinBottomOffset = reader.ReadUE()
	}

	sps.BitDepthLumaMinus8 = reader.ReadUE()
	sps.BitDepthChromaMinus8 = reader.ReadUE()
	sps.Log2MaxPicOrderCntLsbMinus4 = reader.ReadUE()
	sps.SPSSubLayerOrderingInfoPresentFlag = reader.ReadFlag()
	index := sps.SPSMaxSubLayersMinus1
	if sps.SPSSubLayerOrderingInfoPresentFlag {
		index = 0
	}
	sps.SPSMaxDecPicBufferingMinus1 = make([]uint, sps.SPSMaxSubLayersMinus1+1)
	sps.SPSMaxNumReorderPics = make([]uint, sps.SPSMaxSubLayersMinus1+1)
	sps.SPSMaxLatencyIncreasePlus1 = make([]uint, sps.SPSMaxSubLayersMinus1+1)
	for i := index; i <= sps.SPSMaxSubLayersMinus1; i++ {
		sps.SPSMaxDecPicBufferingMinus1[i] = reader.ReadUE()
		sps.SPSMaxNumReorderPics[i] = reader.ReadUE()
		sps.SPSMaxLatencyIncreasePlus1[i] = reader.ReadUE()
	}

	sps.Log2MinLumaCodingBlockSizeMinus3 = reader.ReadUE()
	sps.Log2DiffMaxMinLumaCodingBlockSize = reader.ReadUE()
	sps.Log2MinLumaTransformBlockSizeMinus2 = reader.ReadUE()
	sps.Log2DiffMaxMinLumaTransformBlockSize = reader.ReadUE()
	sps.MaxTransformHierarchyDepthInter = reader.ReadUE()
	sps.MaxTransformHierarchyDepthIntra = reader.ReadUE()
	sps.ScalingListEnabledFlag = reader.ReadFlag()
	if sps.ScalingListEnabledFlag {
		sps.SPSScalingListDataPresentFlag = reader.ReadFlag()
		if sps.SPSScalingListDataPresentFlag {
			sps.scalingListData(reader)
		}
	}

	sps.AmpEnabledFlag = reader.ReadFlag()
	sps.SampleAdaptiveOffsetEnabledFlag = reader.ReadFlag()
	sps.PcmEnabledFlag = reader.ReadFlag()
	if sps.PcmEnabledFlag {
		sps.PcmSampleBitDepthLumaMinus1 = reader.ReadBitsUint8(4)
		sps.PcmSampleBitDepthChromaMinus1 = reader.ReadBitsUint8(4)
		sps.Log2MinPcmLumaCodingBlockSizeMinus3 = reader.ReadUE()
		sps.Log2DiffMaxMinPcmLumaCodingBlockSize = reader.ReadUE()
		sps.PcmLoopFilterDisabledFlag = reader.ReadFlag()
	}
	sps.NumShortTermRefPicSets = reader.ReadUE()
	for i := 0; i < int(sps.NumShortTermRefPicSets) && i < 64; i++ {
		if err := sps.stRefPicSet(reader, i); err != nil {
			logrus.WithField("error", err).Info("parse sps(stRefPicSet) failed")
			return sps, err
		}
	}

	sps.LongTermRefPicsPresentFlag = reader.ReadFlag()
	if sps.LongTermRefPicsPresentFlag {
		sps.NumLongTermRefPicsSps = reader.ReadUE()
		for i := 0; i < int(sps.NumLongTermRefPicsSps); i++ {
			// lt_ref_pic_poc_lsb_sps[ i ]
			reader.ReadBits(int(sps.Log2MaxPicOrderCntLsbMinus4) + 4)
			// used_by_curr_pic_lt_sps_flag[ i ]
			reader.ReadFlag()
		}
	}
	sps.SPSTemporalMvpEnabledFlag = reader.ReadFlag()
	sps.StrongIntraSmoothingEnabledFlag = reader.ReadFlag()
	sps.VUIParametersPresentFlag = reader.ReadFlag()
	sps.VUI = ParseVUI(reader, sps)

	if reader.Error() {
		logrus.Debugf("parse sps failed, the hex string of sps is %s",
			hex.EncodeToString(reader.OriginData()))
		if logrus.GetLevel() == logrus.DebugLevel {
			pretty.Println(sps)
		}
		return sps, errors.New("invalid data")
	}
	return sps, nil
}
