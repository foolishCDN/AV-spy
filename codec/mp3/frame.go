package mp3

// Frame = header + error check + audio data
// audio data = side info + main data
// main data = scale factors + huffman code bits + ancillary data
type Frame struct {
	Header   *FrameHeader
	SideInfo *SideInfo
	MainData *MainData
}
