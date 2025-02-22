package avc

import (
	"github.com/foolishCDN/AV-spy/utils"
)

// SPS Sequence parameter sets
type SPS struct {
	ProfileIdc        uint8 // u(8)
	ConstraintSetFlag uint8 // for constraint_set0_flag to constraint_set5_flag and reserved_zero_2bits
	LevelIdc          uint8 // u(8)
	SeqParameterSetID uint  // ue(v)
	ChromaFormatIdc   uint  // ue(v)

	SeparateColourPlaneFlag         bool   // u(1)
	BitDepthLumaMinus8              uint   // ue(v)
	BitDepthChromaMinus8            uint   // ue(v)
	QPPrimeYZeroTransformBypassFlag bool   // ue(1)
	SetScalingMatrixPresentFlag     bool   // ue(1)
	SeqScalingListPresentFlag       []bool //  ( chroma_format_idc != 3 ) ? 8 : 12
	UseDefaultScalingMatrix4x4Flag  [6]bool
	UseDefaultScalingMatrix8x8Flag  [6]bool
	ScalingList4x4                  [6][16]byte
	ScalingList8x8                  [6][16]byte

	Log2MaxFrameNumMinus4          uint  // ue(v)
	PicOrderCntType                uint  // ue(v)
	Log2MaxPicOrderCntLsbMinus4    uint  // ue(v)
	DeltaPicOrderAlwaysZeroFlag    bool  // u(1)
	OffsetForNonRefPic             int   // se(v)
	OffsetForTopToBottomField      int   // se(v)
	NumRefFramesInPicOrderCntCycle uint  // ue(v)
	OffsetForRefFrame              []int // se(v)

	MaxNumRefFrames                uint // ue(v)
	GapsInFrameNumValueAllowedFlag bool // u(1)
	PicWidthInMbsMinus1            uint // ue(v)
	PicHeightInMapUnitsMinus1      uint // ue(v)
	FrameMbsOnlyFlag               bool // u(1)
	MbAdaptiveFrameFieldFlag       bool // u(1)
	Direct8x8InferenceFlag         bool // u(1)

	FrameCroppingFlag     bool // u(1)
	FrameCropLeftOffset   uint // ue(v)
	FrameCropRightOffset  uint // ue(v)
	FrameCropTopOffset    uint // ue(v)
	FrameCropBottomOffset uint // ue(v)

	VUIParametersPresentFlag bool // u(1)
	VUI                      VUI
}

func (sps *SPS) FPS() float64 {
	return sps.VUI.FPS()
}

func (sps *SPS) Width() int {
	return int(sps.PicWidthInMbsMinus1+1)*16 - int(sps.FrameCropRightOffset)*2 - int(sps.FrameCropLeftOffset)*2
}

func (sps *SPS) Height() int {
	i := 0
	if sps.FrameMbsOnlyFlag {
		i = 1
	}
	return int(2-i)*int(sps.PicHeightInMapUnitsMinus1+1)*16 - int(sps.FrameCropBottomOffset)*2 - int(sps.FrameCropTopOffset)*2
}

func ParseSPS(reader *utils.BitReader) SPS {
	sps := SPS{}

	sps.ProfileIdc = reader.ReadBitsUint8(8)
	sps.ConstraintSetFlag = reader.ReadBitsUint8(8)
	sps.LevelIdc = reader.ReadBitsUint8(8)
	sps.SeqParameterSetID = reader.ReadUE()

	switch sps.ProfileIdc {
	case 100, 110, 122, 244, 44, 83, 86, 118, 128, 138, 139, 134, 135:
		sps.ChromaFormatIdc = reader.ReadUE()
		if sps.ChromaFormatIdc == 3 {
			sps.SeparateColourPlaneFlag = reader.ReadFlag()
		}
		sps.BitDepthLumaMinus8 = reader.ReadUE()
		sps.BitDepthChromaMinus8 = reader.ReadUE()
		sps.QPPrimeYZeroTransformBypassFlag = reader.ReadFlag()
		sps.SetScalingMatrixPresentFlag = reader.ReadFlag()
		if sps.SetScalingMatrixPresentFlag {
			n := 12
			if sps.ChromaFormatIdc != 3 {
				n = 8
			}
			sps.SeqScalingListPresentFlag = make([]bool, n)
			for i := 0; i < n; i++ {
				sps.SeqScalingListPresentFlag[i] = reader.ReadFlag()
				if sps.SeqScalingListPresentFlag[i] {
					if i < 6 {
						ScalingList(reader, sps.ScalingList4x4[i][:], &sps.UseDefaultScalingMatrix4x4Flag[i])
					} else {
						ScalingList(reader, sps.ScalingList8x8[i-6][:], &sps.UseDefaultScalingMatrix8x8Flag[i])
					}
				}
			}
		}
	}
	sps.Log2MaxFrameNumMinus4 = reader.ReadUE()
	sps.PicOrderCntType = reader.ReadUE()
	if sps.PicOrderCntType == 0 {
		sps.Log2MaxPicOrderCntLsbMinus4 = reader.ReadUE()
	} else if sps.PicOrderCntType == 1 {
		sps.DeltaPicOrderAlwaysZeroFlag = reader.ReadFlag()
		sps.OffsetForNonRefPic = reader.ReadSE()
		sps.OffsetForTopToBottomField = reader.ReadSE()
		sps.NumRefFramesInPicOrderCntCycle = reader.ReadUE()
		sps.OffsetForRefFrame = make([]int, sps.NumRefFramesInPicOrderCntCycle)
		for i := 0; i < int(sps.NumRefFramesInPicOrderCntCycle); i++ {
			sps.OffsetForRefFrame[i] = reader.ReadSE()
		}
	}

	sps.MaxNumRefFrames = reader.ReadUE()
	sps.GapsInFrameNumValueAllowedFlag = reader.ReadFlag()
	sps.PicWidthInMbsMinus1 = reader.ReadUE()
	sps.PicHeightInMapUnitsMinus1 = reader.ReadUE()
	sps.FrameMbsOnlyFlag = reader.ReadFlag()
	if !sps.FrameMbsOnlyFlag {
		sps.MbAdaptiveFrameFieldFlag = reader.ReadFlag()
	}

	sps.Direct8x8InferenceFlag = reader.ReadFlag()
	sps.FrameCroppingFlag = reader.ReadFlag()
	if sps.FrameCroppingFlag {
		sps.FrameCropLeftOffset = reader.ReadUE()
		sps.FrameCropRightOffset = reader.ReadUE()
		sps.FrameCropTopOffset = reader.ReadUE()
		sps.FrameCropBottomOffset = reader.ReadUE()
	}

	sps.VUIParametersPresentFlag = reader.ReadFlag()
	if sps.VUIParametersPresentFlag {
		sps.VUI = ParseVUI(reader)
	}
	return sps
}

func ScalingList(reader *utils.BitReader, scalingList []byte, useDefaultScalingMatrixFlag *bool) {
	lastScale := 8
	nextScale := 8
	for i := 0; i < len(scalingList); i++ {
		if nextScale != 0 {
			deltaScale := int(reader.ReadSE())
			nextScale = (lastScale + deltaScale + 256) % 256
			*useDefaultScalingMatrixFlag = i == 0 && nextScale == 0
		}
		if nextScale == 0 {
			scalingList[i] = byte(lastScale)
		} else {
			scalingList[i] = byte(nextScale)
		}
		lastScale = int(scalingList[i])
	}
}
