package tests

import (
	"Torrent-Client/client"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHandshake(t *testing.T) {

	infoHash := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	peerID := [20]byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	hShake := client.NewHandshake(peerID, infoHash)

	expectedHandshake := client.Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	assert.Equal(t, expectedHandshake, *hShake)
}

func TestHandshake_ReadResponse(t *testing.T) {

	input := []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	output := &client.Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
		PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}

	reader := bytes.NewReader(input)

	h, err := client.ReadMessage(reader)

	assert.Nil(t, err)
	assert.Equal(t, output, h)

	input = []byte{}
	reader = bytes.NewReader(input)

	h, err = client.ReadMessage(reader)

	assert.Nil(t, h)
	assert.NotNil(t, err)

	input = []byte{0}
	reader = bytes.NewReader(input)

	h, err = client.ReadMessage(reader)

	assert.Nil(t, h)
	assert.NotNil(t, err)

	input = []byte{14, 20, 13, 32}
	reader = bytes.NewReader(input)

	h, err = client.ReadMessage(reader)

	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestHandshake_Serialize(t *testing.T) {

	input := &client.Handshake{
		Pstr:     "Different protocol",
		InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
		PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}

	output := []byte{32, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 44, 32, 98, 117, 116, 32, 99, 111, 111, 108, 101, 114, 63, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	assert.NotEqual(t, output, input.Serialize())

	input = &client.Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
		PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}

	output = []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	assert.Equal(t, output, input.Serialize())
}
