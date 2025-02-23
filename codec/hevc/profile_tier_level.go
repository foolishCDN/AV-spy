package hevc

import "github.com/foolishCDN/AV-spy/utils"

type ProfileTierLevel struct {
	GeneralProfileSpace             uint8
	GeneralTierFlag                 bool
	GeneralProfileIDC               uint8
	GeneralProfileCompatibilityFlag uint32
	GeneralConstraintFlags          uint64
	GeneralLevelIdc                 uint8

	SubLayerProfilePresentFlag       []bool
	SubLayerLevelPresentFlag         []bool
	SubLayerProfileSpace             []uint8
	SubLayerTierFlag                 []bool
	SubLayerProfileIdc               []uint8
	SubLayerProfileCompatibilityFlag []uint32
	SubLayerConstraintFlags          []uint64
	SubLayerLevelIdc                 []uint8
}

func ParseProfileTierLevel(reader *utils.BitReader, profilePresentFlag bool, maxNumSubLayersMinus1 uint8) *ProfileTierLevel {
	ptl := new(ProfileTierLevel)
	if profilePresentFlag {
		ptl.GeneralProfileSpace = reader.ReadBitsUint8(2)
		ptl.GeneralTierFlag = reader.ReadFlag()
		ptl.GeneralProfileIDC = reader.ReadBitsUint8(5)
		ptl.GeneralProfileCompatibilityFlag = reader.ReadBitsUint32(32)
		ptl.GeneralConstraintFlags = reader.ReadBitsUint64(48)
	}
	ptl.GeneralLevelIdc = reader.ReadBitsUint8(8)

	ptl.SubLayerProfilePresentFlag = make([]bool, maxNumSubLayersMinus1)
	ptl.SubLayerLevelPresentFlag = make([]bool, maxNumSubLayersMinus1)
	for i := 0; i < int(maxNumSubLayersMinus1); i++ {
		ptl.SubLayerProfilePresentFlag[i] = reader.ReadFlag()
		ptl.SubLayerLevelPresentFlag[i] = reader.ReadFlag()
	}

	// align to byte
	if maxNumSubLayersMinus1 > 0 {
		for i := maxNumSubLayersMinus1; i < 8; i++ {
			reader.ReadBits(2) // reserved_zero_2bits
		}
	}

	for i := 0; i < int(maxNumSubLayersMinus1); i++ {
		if ptl.SubLayerProfilePresentFlag[i] {
			ptl.SubLayerProfileSpace = append(ptl.SubLayerProfileSpace, reader.ReadBitsUint8(2))
			ptl.SubLayerTierFlag = append(ptl.SubLayerTierFlag, reader.ReadFlag())
			ptl.SubLayerProfileIdc = append(ptl.SubLayerProfileIdc, reader.ReadBitsUint8(5))
			ptl.SubLayerProfileCompatibilityFlag = append(ptl.SubLayerProfileCompatibilityFlag, reader.ReadBitsUint32(32))
			ptl.SubLayerConstraintFlags = append(ptl.SubLayerConstraintFlags, reader.ReadBitsUint64(48))
		}
		if ptl.SubLayerLevelPresentFlag[i] {
			ptl.SubLayerLevelIdc = append(ptl.SubLayerLevelIdc, reader.ReadBitsUint8(8))
		}
	}

	return ptl
}
