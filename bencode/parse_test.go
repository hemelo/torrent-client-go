package bencode

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"testing"
)

func loadTorrent() string {
	// Read the file content
	data, err := os.ReadFile("../demo.torrent")

	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Convert the content to a string
	return string(data)
}

func TestParse(t *testing.T) {
	// Fictional bencoded torrent data
	torrentData := loadTorrent()

	reader := io.Reader(bufio.NewReader(bytes.NewReader([]byte(torrentData))))
	result, err := Parse(reader)

	if err != nil {
		t.Fatalf("Failed to parse bencoded data: %v", err)
	}

	// Check the parsed result
	if result.Type != DictType {
		t.Fatalf("Expected DictType, got %v", result.Type)
	}

	// Check the announce URL
	announce, ok := result.Dict["announce"]
	if !ok || announce.Type != StringType || announce.Str != "http://tracker.example.com/announce" {
		t.Fatalf("Expected announce URL to be 'http://tracker.example.com/announce', got %v", announce)
	}

	// Check the created by field
	createdBy, ok := result.Dict["created by"]
	if !ok || createdBy.Type != StringType || createdBy.Str != "example client" {
		t.Fatalf("Expected created by to be 'example client', got %v", createdBy)
	}

	// Check the creation date
	creationDate, ok := result.Dict["creation date"]
	if !ok || creationDate.Type != IntegerType || creationDate.Int != 1633036800 {
		t.Fatalf("Expected creation date to be 1633036800, got %v", creationDate)
	}

	// Check the info dictionary
	info, ok := result.Dict["info"]
	if !ok || info.Type != DictType {
		t.Fatalf("Expected info to be a dictionary, got %v", info)
	}

	// Check the name field in the info dictionary
	name, ok := info.Dict["name"]
	if !ok || name.Type != StringType || name.Str != "example.torrent" {
		t.Fatalf("Expected name to be 'example.torrent', got %v", name)
	}

	// Check the piece length
	pieceLength, ok := info.Dict["piece length"]
	if !ok || pieceLength.Type != IntegerType || pieceLength.Int != 524288 {
		t.Fatalf("Expected piece length to be 524288, got %v", pieceLength)
	}

	// Check the pieces field
	pieces, ok := info.Dict["pieces"]
	if !ok || pieces.Type != StringType || pieces.Str != "abcdefghij0123456789" {
		t.Fatalf("Expected pieces to be 'abcdefghij0123456789', got %v", pieces)
	}
}
