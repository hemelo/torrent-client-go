package torrent

import (
	"Torrent-Client/bencode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewTorrentFrom(t *testing.T) {
	_, err := NewTorrentFrom("../test_data/demo.torrent")
	require.NoError(t, err, "Failed to load torrent from file")
}

func TestMissingFile(t *testing.T) {
	_, err := NewTorrentFrom("../test_data/missing.torrent")
	require.Error(t, err, "Expected error when file is missing")
}

func TestWithoutParse(t *testing.T) {
	tests := map[string]struct {
		input  *bencode.BencodeValue
		output TorrentFile
		fails  bool
	}{
		"correct conversion": {
			input: &bencode.BencodeValue{
				Dict: map[string]bencode.BencodeValue{
					"comment":       {Str: "\"Debian CD from cdimage.debian.org\"", Type: bencode.StringType},
					"announce":      {Str: "http://bttracker.debian.org:6969/announce", Type: bencode.StringType},
					"created by":    {Str: "mktorrent 1.1", Type: bencode.StringType},
					"creation date": {Int: 1731156219, Type: bencode.IntegerType},
					"info": bencode.BencodeValue{
						Dict: map[string]bencode.BencodeValue{
							"pieces":       {Str: "1234567890abcdefghijabcdefghij1234567890", Type: bencode.StringType},
							"piece length": {Int: 262144, Type: bencode.IntegerType},
							"length":       {Int: 351272960, Type: bencode.IntegerType},
							"name":         {Str: "debian-10.2.0-amd64-netinst.iso", Type: bencode.StringType},
						},
						Type: bencode.DictType,
					},
				},
				Type: bencode.DictType,
			},
			output: TorrentFile{
				Comment:      "\"Debian CD from cdimage.debian.org\"",
				Announce:     "http://bttracker.debian.org:6969/announce",
				CreatedBy:    "mktorrent 1.1",
				CreationDate: 1731156219,
				InfoHash:     [20]byte{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
				PiecesHash: [][20]byte{
					{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
					{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
				},
				PieceLength: 262144,
				Length:      351272960,
				Name:        "debian-10.2.0-amd64-netinst.iso",
			},
			fails: false,
		},
	}

	for _, test := range tests {
		to, err := BencodeToTorrentFile(*test.input, BencodeToTorrentFileOpts{from: "test"})

		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, test.output, to)
	}
}
