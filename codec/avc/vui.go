package avc

import (
	"github.com/foolishCDN/AV-spy/utils"
)

const VUIExtendedSar = 255

// VUI Video usability information
type VUI struct {
	AspectRatioInfoPresentFlag bool
	AspectRatioIdc             uint8
	SarWidth                   uint16
	SarHeight                  uint16

	OverscanInfoPresentFlag      bool
	OverscanAppropriateFlag      bool
	VideoSignalTypePresentFlag   bool
	VideoFormat                  uint8
	VideoFullRangeFlag           bool
	ColourDescriptionPresentFlag bool
	ColourPrimaries              uint8
	TransferCharacteristics      uint8
	MatrixCoefficients           uint8

	ChromaLocInfoPresentFlag       bool
	ChromaSampleLocTypeTopField    uint
	ChromaSampleLocTypeBottomField uint

	TimingInfoPresentFlag bool
	NumUnitsInTick        uint32
	TimeScale             uint32
	FixedFrameRateFlag    bool

	NalHrdParametersPresentFlag bool
	NalHrd                      HRD
	VclHrdParametersPresentFlag bool
	VclHrd                      HRD
	LowDelayHrdFlag             bool
	PicStructPresentFlag        bool

	BitstreamRestrictionFlag           bool
	MotionVectorsOverPicBoundariesFlag bool
	MaxBytesPerPicDenom                uint
	MaxBitsPerMbDenom                  uint
	Log2MaxMvLengthHorizontal          uint
	Log2MaxMvLengthVertical            uint
	MaxNumReorderFrames                uint
	MaxDecFrameBuffering               uint
}

func (vui *VUI) FPS() float64 {
	return float64(vui.TimeScale) / float64(vui.NumUnitsInTick) / 2.0
}

func ParseVUI(reader *utils.BitReader) VUI {
	vui := VUI{}
	vui.AspectRatioInfoPresentFlag = reader.ReadFlag()
	if vui.AspectRatioInfoPresentFlag {
		vui.AspectRatioIdc = reader.ReadBitsUint8(8)
		if vui.AspectRatioIdc == VUIExtendedSar {
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
			vui.MatrixCoefficients = reader.ReadBitsUint8(8)
		}
	}

	vui.ChromaLocInfoPresentFlag = reader.ReadFlag()
	if vui.ChromaLocInfoPresentFlag {
		vui.ChromaSampleLocTypeTopField = reader.ReadUE()
		vui.ChromaSampleLocTypeBottomField = reader.ReadUE()
	}

	vui.TimingInfoPresentFlag = reader.ReadFlag()
	if vui.TimingInfoPresentFlag {
		vui.NumUnitsInTick = reader.ReadBitsUint32(32)
		vui.TimeScale = reader.ReadBitsUint32(32)
		vui.FixedFrameRateFlag = reader.ReadFlag()
	}

	vui.NalHrdParametersPresentFlag = reader.ReadFlag()
	if vui.NalHrdParametersPresentFlag {
		vui.NalHrd = ParseHRD(reader)
	}

	vui.VclHrdParametersPresentFlag = reader.ReadFlag()
	if vui.VclHrdParametersPresentFlag {
		vui.VclHrd = ParseHRD(reader)
	}

	if vui.NalHrdParametersPresentFlag || vui.VclHrdParametersPresentFlag {
		vui.LowDelayHrdFlag = reader.ReadFlag()
	}

	vui.PicStructPresentFlag = reader.ReadFlag()
	vui.BitstreamRestrictionFlag = reader.ReadFlag()
	if vui.BitstreamRestrictionFlag {
		vui.MotionVectorsOverPicBoundariesFlag = reader.ReadFlag()
		vui.MaxBytesPerPicDenom = reader.ReadUE()
		vui.MaxBitsPerMbDenom = reader.ReadUE()
		vui.Log2MaxMvLengthHorizontal = reader.ReadUE()
		vui.Log2MaxMvLengthVertical = reader.ReadUE()
		vui.MaxNumReorderFrames = reader.ReadUE()
		vui.MaxDecFrameBuffering = reader.ReadUE()
	}
	return vui
}
