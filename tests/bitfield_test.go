package tests

import (
	"Torrent-Client/client"
	"testing"
)

func TestBitfield_HasPiece(t *testing.T) {
	bf := client.Bitfield{0x80} // 10000000

	tests := []struct {
		index    int
		expected bool
	}{
		{0, true},
		{1, false},
		{7, false},
		{8, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := bf.HasPiece(tt.index); got != tt.expected {
				t.Errorf("HasPiece(%d) = %v, expected %v", tt.index, got, tt.expected)
			}
		})
	}
}

func TestBitfield_SetPiece(t *testing.T) {
	bf := client.Bitfield{0x00} // 00000000

	bf.SetPiece(0)
	if bf[0] != 0x80 { // 10000000
		t.Errorf("SetPiece(0) failed, got %08b", bf[0])
	}

	bf.SetPiece(7)
	if bf[0] != 0x81 { // 10000001
		t.Errorf("SetPiece(7) failed, got %08b", bf[0])
	}

	bf.SetPiece(3)
	if bf[0] != 0x89 { // 10001001
		t.Errorf("SetPiece(3) failed, got %08b", bf[0])
	}
}
