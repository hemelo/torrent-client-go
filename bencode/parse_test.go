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
	data, err := os.ReadFile("../test_data/demo.torrent")

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
	if !ok || announce.Type != StringType || announce.Str != "http://bttracker.debian.org:6969/announce" {
		t.Fatalf("Expected announce URL to be 'http://tracker.example.com/announce', got %v", announce.Str)
	}

	// Check comment field
	comment, ok := result.Dict["comment"]
	if !ok || comment.Type != StringType || comment.Str != "\"Debian CD from cdimage.debian.org\"" {
		t.Fatalf("Expected comment to be '\"Debian CD from cdimage.debian.org\"', got %v", comment.Str)
	}

	// Check the created by field
	createdBy, ok := result.Dict["created by"]
	if !ok || createdBy.Type != StringType || createdBy.Str != "mktorrent 1.1" {
		t.Fatalf("Expected created by to be 'example client', got %v", createdBy.Str)
	}

	// Check the creation date
	creationDate, ok := result.Dict["creation date"]
	if !ok || creationDate.Type != IntegerType || creationDate.Int != 1731156219 {
		t.Fatalf("Expected creation date to be 1633036800, got %v", creationDate.Int)
	}

	// Check the info dictionary
	info, ok := result.Dict["info"]
	if !ok || info.Type != DictType {
		t.Fatalf("Expected info to be a dictionary, got %v", info.Type)
	}

	// Check the name field in the info dictionary
	name, ok := info.Dict["name"]
	if !ok || name.Type != StringType || name.Str != "debian-12.8.0-amd64-netinst.iso" {
		t.Fatalf("Expected name to be 'debian-12.8.0-amd64-netinst.iso', got %v", name.Str)
	}

	// Check the piece length
	pieceLength, ok := info.Dict["piece length"]
	if !ok || pieceLength.Type != IntegerType || pieceLength.Int != 262144 {
		t.Fatalf("Expected piece length to be 262144, got %v", pieceLength.Int)
	}

	// Check the pieces field
	pieces, ok := info.Dict["pieces"]
	if !ok || pieces.Type != StringType {
		t.Fatalf("Expected pieces to be a string, got %v", pieces.Type)
	}

}
