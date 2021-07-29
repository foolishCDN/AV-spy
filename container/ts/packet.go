package ts

type Packet struct {
	Header  *PacketHeader
	Payload []byte
}

type PacketHeader struct {
	TransportErrorIndicator    bool
	PayloadUnitStartIndicator  bool
	TransportPriority          bool
	PID                        uint16
	TransportScramblingControl uint8
	AdaptationFieldControl     bool
	ContinuityCounter          uint8
}
