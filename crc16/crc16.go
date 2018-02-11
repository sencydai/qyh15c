package crc16

const (
	POLYNOMIAL int64 = 0x1021
)

var (
	CRCTable = makeCRCTable()
	DropBits = []int64{0xffffffff, 0xfffffffe, 0xfffffffc, 0xfffffff8,
		0xfffffff0, 0xffffffe0, 0xffffffc0, 0xffffff80,
		0xffffff00, 0xfffffe00, 0xfffffc00, 0xfffff800,
		0xfffff000, 0xffffe000, 0xffffc000, 0xffff8000}
)

func cRCBitReflect(input, bitCount int64) int64 {
	var out int64
	var x int64

	bitCount--
	for i := int64(0); i <= bitCount; i++ {
		x = bitCount - i
		if (input & 1) != 0 {
			out |= (1 << uint32(x)) & DropBits[x]
		}
		input = (input >> 1) & 0x7fffffff
	}
	return out
}

func makeCRCTable() []int64 {
	var c int64
	crcTable := make([]int64, 256)
	for i := int64(0); i < 256; i++ {
		c = (i << 8) & 0xffffff00
		for j := int64(0); j < 8; j++ {
			if (c & 0x8000) != 0 {
				c = ((c << 1) & 0xfffffffe) ^ POLYNOMIAL
			} else {
				c = (c << 1) & 0xfffffffe
			}
		}
		crcTable[i] = c
	}
	return crcTable
}

func Update(data []byte, offset, length int64) int16 {
	var c int64
	var index int64
	if length == 0 {
		length = int64(len(data))
	}
	pos := offset
	for i := offset; i < length; i++ {
		b := data[pos]
		pos++
		index = (cRCBitReflect(int64(b), 8) & 0xff) ^ ((c >> 8) & 0xffffff)
		index &= 0xff
		c = CRCTable[index] ^ ((c << 8) & 0xffffff00)
	}
	return int16((cRCBitReflect(c, 16) ^ 0) & 0xffff)
}
