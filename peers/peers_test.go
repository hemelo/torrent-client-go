package peers

import "testing"

func TestDecodePeers(t *testing.T) {
	// Test case where the peers data is empty
	peers, err := DecodePeers([]byte{})
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	// Test case where the peers data is malformed
	peers, err = DecodePeers([]byte{1, 2, 3, 4, 5})
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	// Test case where the peers data is correct
	peers, err = DecodePeers([]byte{192, 168, 1, 1, 0, 80})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(peers))
	}

	if peers[0].IP.String() != "192.168.1.1" {
		t.Fatalf("Expected IP to be 192.168.1.1")
	}

	if peers[0].Port != 80 {
		t.Fatalf("Expected port to be 80")
	}
}
