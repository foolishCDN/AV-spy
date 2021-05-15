package wav

// For others, see https://www.recordingblogs.com/wiki/format-chunk-of-a-wave-file
const (
	FormatPCM       = 1
	FormatIEEEFloat = 3
)

type Format struct {
	Format        uint16 // see above
	NumOfChannels uint16
	SampleRate    uint32 // 8000, 44100, etc
	ByteRate      uint32 // sampleRate * numOfChannels * bitPerSample/8
	BlocKAlign    uint16 // data block size (bytes), numOfChannels * bitPerSample/8
	BitPerSample  uint16

	ExtraFormatSize uint16 //  if PCM, then doesn't exist
	ExtraFormat     []byte
}
