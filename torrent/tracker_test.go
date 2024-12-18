package torrent

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/url"
	"testing"
)

func TestBuildTrackerUrl(t *testing.T) {

	var piecesHashes [][20]byte
	var infoHash [20]byte
	var peerID [20]byte

	for i := 0; i < 4; i++ {
		var pieceHash [20]byte

		for j := 0; j < 20; j++ {
			pieceHash[j] = byte(rand.Intn(100))
		}

		piecesHashes = append(piecesHashes, pieceHash)
	}

	for i := 0; i < 20; i++ {
		peerID[i] = byte(rand.Intn(100)) // Random integer between 0 and 99
	}

	for i := 0; i < 20; i++ {
		infoHash[i] = byte(rand.Intn(100))
	}

	torrent := TorrentFile{
		PieceLength: 262144,
		Length:      351272960,
		Name:        "debian-10.2.0-amd64-netinst.iso",
		Announce:    "http://bttracker.debian.org:6969/announce",
		InfoHash:    infoHash,
		PiecesHash:  piecesHashes,
	}

	port := uint16(rand.Intn(65535)) // Random integer between 0 and 65534

	expected := fmt.Sprintf("http://bttracker.debian.org:6969/announce?compact=1&downloaded=0&info_hash=%s&left=351272960&peer_id=%s&port=%d&uploaded=0",
		url.QueryEscape(string(infoHash[:])),
		url.QueryEscape(string(peerID[:])),
		port,
	)

	trackerURL, err := torrent.buildTrackerUrl(peerID, port)

	assert.Nil(t, err, "Failed to build tracker URL")
	assert.Equal(t, expected, trackerURL, "Unexpected tracker URL")
}
