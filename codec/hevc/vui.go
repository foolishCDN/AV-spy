package hevc

import "github.com/foolishCDN/AV-spy/utils"

const ExtendedSAR = 255

type VUI struct {
	AspectRatioInfoPresentFlag bool
	AspectRatioIdc             uint8
	SarWidth                   uint16
	SarHeight                  uint16

	OverscanInfoPresentFlag bool
	OverscanAppropriateFlag bool

	VideoSignalTypePresentFlag   bool
	VideoFormat                  uint8
	VideoFullRangeFlag           bool
	ColourDescriptionPresentFlag bool
	ColourPrimaries              uint8
	TransferCharacteristics      uint8
	MatrixCoeffs                 uint8

	ChromaLocInfoPresentFlag       bool
	ChromaSampleLocTypeTopField    uint
	ChromaSampleLocTypeBottomField uint

	NeutralChromaIndicationFlag bool
	FieldSeqFlag                bool
	FrameFieldInfoPresentFlag   bool

	DefaultDisplayWindowFlag bool
	DefDispWinLeftOffset     uint
	DefDispWinRightOffset    uint
	DefDispWinTopOffset      uint
	DefDispWinBottomOffset   uint

	VUITimingInfoPresentFlag       bool
	VUINumUnitsInTick              uint32
	VUITimeScale                   uint32
	VUIPocProportionalToTimingFlag bool
	VUINumTicksPocDiffOneMinus1    uint
	VUIHrdParametersPresentFlag    bool
	HRD                            *HRD

	BitstreamRestrictionFlag           bool
	TilesFixedStructureFlag            bool
	MotionVectorsOverPicBoundariesFlag bool
	RestrictedRefPicListsFlag          bool
	MinSpatialSegmentationIdc          uint
	MaxBytesPerPicDenom                uint
	MaxBitsPerMinCuDenom               uint
	Log2MaxMvLengthHorizontal          uint
	Log2MaxMvLengthVertical            uint
}

func (vui *VUI) FPS() float64 {
	return float64(vui.VUITimeScale) / float64(vui.VUINumUnitsInTick)
}

func ParseVUI(reader *utils.BitReader, sps *SPS) *VUI {
	vui := new(VUI)
	vui.AspectRatioInfoPresentFlag = reader.ReadFlag()
	if vui.AspectRatioInfoPresentFlag {
		vui.AspectRatioIdc = reader.ReadBitsUint8(8)
		if vui.AspectRatioIdc == ExtendedSAR {
			vui.SarWidth = reader.ReadBitsUint16(16)
			vui.SarHeight = reader.ReadBitsUint16(16)
		}
	}

	vui.OverscanInfoPresentFlag = reader.ReadFlag()
	if vui.OverscanInfoPresentFlag {
		vui.OverscanAppropriateFlag = reader.ReadFlag()
	}

	vui.VideoSignalTypePresentFlag = reader.ReadFlag()
	if vui.VideoSignalTypePresentFlag {
		vui.VideoFormat = reader.ReadBitsUint8(3)
		vui.VideoFullRangeFlag = reader.ReadFlag()
		vui.ColourDescriptionPresentFlag = reader.ReadFlag()
		if vui.ColourDescriptionPresentFlag {
			vui.ColourPrimaries = reader.ReadBitsUint8(8)
			vui.TransferCharacteristics = reader.ReadBitsUint8(8)
			vui.MatrixCoeffs = reader.ReadBitsUint8(8)
		}
	}
	vui.ChromaLocInfoPresentFlag = reader.ReadFlag()
	if vui.ChromaLocInfoPresentFlag {
		vui.ChromaSampleLocTypeTopField = reader.ReadUE()
		vui.ChromaSampleLocTypeBottomField = reader.ReadUE()
	}
	vui.NeutralChromaIndicationFlag = reader.ReadFlag()
	vui.FieldSeqFlag = reader.ReadFlag()
	vui.FrameFieldInfoPresentFlag = reader.ReadFlag()
	vui.DefaultDisplayWindowFlag = reader.ReadFlag()
	if vui.DefaultDisplayWindowFlag {
		vui.DefDispWinLeftOffset = reader.ReadUE()
		vui.DefDispWinRightOffset = reader.ReadUE()
		vui.DefDispWinTopOffset = reader.ReadUE()
		vui.DefDispWinBottomOffset = reader.ReadUE()
	}
	vui.VUITimingInfoPresentFlag = reader.ReadFlag()
	if vui.VUITimingInfoPresentFlag {
		vui.VUINumUnitsInTick = reader.ReadBitsUint32(32)
		vui.VUITimeScale = reader.ReadBitsUint32(32)
		vui.VUIPocProportionalToTimingFlag = reader.ReadFlag()
		if vui.VUIPocProportionalToTimingFlag {
			vui.VUINumTicksPocDiffOneMinus1 = reader.ReadUE()
		}
		vui.VUIHrdParametersPresentFlag = reader.ReadFlag()
		if vui.VUIHrdParametersPresentFlag {
			vui.HRD = ParseHRD(reader, true, sps.SPSMaxSubLayersMinus1)
		}
	}
	vui.BitstreamRestrictionFlag = reader.ReadFlag()
	if vui.BitstreamRestrictionFlag {
		vui.TilesFixedStructureFlag = reader.ReadFlag()
		vui.MotionVectorsOverPicBoundariesFlag = reader.ReadFlag()
		vui.RestrictedRefPicListsFlag = reader.ReadFlag()
		vui.MinSpatialSegmentationIdc = reader.ReadUE()
		vui.MaxBytesPerPicDenom = reader.ReadUE()
		vui.MaxBitsPerMinCuDenom = reader.ReadUE()
		vui.Log2MaxMvLengthHorizontal = reader.ReadUE()
		vui.Log2MaxMvLengthVertical = reader.ReadUE()
	}
	return vui
}
