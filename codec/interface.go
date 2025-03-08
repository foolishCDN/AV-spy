package codec

type SPS interface {
	Width() int
	Height() int
	FPS() float64
}

type NALUType interface {
	String() string
}
