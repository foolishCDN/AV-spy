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

type MIDISample struct {
	Manufacturer        uint16
	Product             uint16
	SamplePeriod        uint16
	MIDIUnityNote       uint16
	MIDIPitchFraction   uint16
	SMPTEFormat         uint16
	SMPTEOffset         uint16
	NumOfSampleLoops    uint16
	SampleData          uint16
	SampleLoops         []MIDISampleLoop
	SamplerSpecificData []byte
}

type MIDISampleLoop struct {
	ID                   uint16
	Type                 uint16
	Start                uint16
	End                  uint16
	Fraction             uint16
	NumOfTimesToPlayLoop uint16
}
