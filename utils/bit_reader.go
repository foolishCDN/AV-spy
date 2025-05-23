package utils

func NewBitReader(data []byte) *BitReader {
	return &BitReader{
		data: data,
	}
}

type BitReader struct {
	data        []byte
	offsetBits  uint8
	offsetBytes int
}

func (reader *BitReader) LastData() []byte {
	return reader.data[reader.offsetBytes:]
}

func (reader *BitReader) OriginData() []byte {
	return reader.data
}

func (reader *BitReader) move() {
	if reader.offsetBits++; reader.offsetBits >= 8 {
		reader.offsetBits = 0
		reader.offsetBytes++

		if reader.offsetBytes < len(reader.data) {
			// 0x00 0x00 0x03 0xXX -> skip 0x03
			if reader.offsetBytes >= 2 {
				if BigEndianUint24(reader.data[reader.offsetBytes-2:reader.offsetBytes+1]) == 0x03 {
					reader.offsetBytes++
				}
			}
		}
	}
}
func (reader *BitReader) Error() bool {
	return reader.offsetBytes*8+int(reader.offsetBits) >= len(reader.data)*8
}

func (reader *BitReader) PeekBit() byte {
	if reader.Error() {
		return 0
	}
	return (reader.data[reader.offsetBytes] >> (7 - reader.offsetBits)) & 0x01
}

func (reader *BitReader) ReadBit() byte {
	b := reader.PeekBit()
	reader.move()
	return b
}

func (reader *BitReader) ReadFlag() bool {
	return reader.ReadBit() == 1
}

func (reader *BitReader) ReadBits(n int) uint64 {
	var v uint64
	for i := 0; i < n && i < 64; i++ {
		b := reader.ReadBit()
		v = (v << 1) | uint64(b)
	}
	return v
}

func (reader *BitReader) ReadBitsUint8(n int) uint8 {
	return uint8(reader.ReadBits(n))
}

func (reader *BitReader) Read8BitsUntilNot0xFF() int {
	var res int
	got := reader.ReadBitsUint8(8)
	for got == 0xFF {
		res += 255
		got = reader.ReadBitsUint8(8)
	}
	res += int(got)
	return res
}

func (reader *BitReader) ReadBitsUint16(n int) uint16 {
	return uint16(reader.ReadBits(n))
}

func (reader *BitReader) ReadBitsUint32(n int) uint32 {
	return uint32(reader.ReadBits(n))
}

func (reader *BitReader) ReadBitsUint64(n int) uint64 {
	return uint64(reader.ReadBits(n))
}

// ReadUE read an exponential-Golomb code to uint
func (reader *BitReader) ReadUE() uint {
	leadingZeroBits := 0

	// read the leadingZeroBits, note we already read the first 0b1
	for {
		b := reader.ReadBit()
		if reader.Error() {
			return 0
		}
		if !(b == 0 && leadingZeroBits < 32) {
			break
		}
		leadingZeroBits++
	}
	if leadingZeroBits == 0 {
		return 0
	}
	// read the lastBits
	v := reader.ReadBits(leadingZeroBits)
	if reader.Error() {
		return 0
	}
	// 1 << leadingZeroBits (for the first 0b1) + lastBits - 1
	v += (1 << uint(leadingZeroBits)) - 1
	return uint(v)
}

func (reader *BitReader) ReadSE() int {
	v := reader.ReadUE()
	if v%2 == 0 {
		return -1 * int((v+1)/2)
	}
	return int((v + 1) / 2)
}
