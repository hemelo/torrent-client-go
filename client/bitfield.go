package client

// Bitfield is a data structure that peers use to efficiently encode which pieces they are able to send us
// To check the pieces we just need to check the bits in the bitfield
// By working with bits instead of bytes makes it more efficient
type Bitfield []byte

func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	bitIndex := 7 - (index % 8)

	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}

	return bf[byteIndex]>>(bitIndex)&1 != 0
}

func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	bitIndex := 7 - (index % 8)

	if byteIndex < 0 || byteIndex >= len(bf) {
		return
	}

	bf[byteIndex] |= 1 << bitIndex
}
