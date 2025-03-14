package avc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitNALUsAnnexB(t *testing.T) {
	h264 := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x4d, 0x40, 0x1f, 0xe8, 0x80, 0x6c, 0x1e, 0xf3, 0x78, 0x08, 0x80, 0x00, 0x01, 0xf4, 0x80, 0x00, 0x75, 0x30, 0x07, 0x8c, 0x18, 0x89,
		0x00, 0x00, 0x00, 0x01, 0x68, 0xeb, 0xaf, 0x20,
		0x00, 0x00, 0x00, 0x01, 0x65, 0x88, 0x84, 0x01, 0x7f, 0xec, 0x05, 0x17,
		0x00, 0x00, 0x00, 0x00, 0x01, 0xab,
		0x00, 0x00, 0x01,
	}
	nalus := SplitNALUsAnnexB(h264)
	assert.Equal(t, 4, len(nalus))
}
