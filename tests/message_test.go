package tests

import (
	"Torrent-Client/client"
	"bytes"
	"testing"
)

func TestMessage_Serialize(t *testing.T) {
	tests := []struct {
		name     string
		message  client.Message
		expected []byte
	}{
		{
			name: "keep-alive",
			message: client.Message{
				ID:      client.MessageKeepAlive,
				Payload: nil,
			},
			expected: []byte{0x0, 0x0, 0x0, 0x0},
		},
		{
			name: "choke",
			message: client.Message{
				ID:      client.MessageChoke,
				Payload: nil,
			},
			expected: []byte{0x0, 0x0, 0x0, 0x1, 0x0},
		},
		{
			name: "unchoke",
			message: client.Message{
				ID:      client.MessageUnchoke,
				Payload: nil,
			},
			expected: []byte{0x0, 0x0, 0x0, 0x1, 0x1},
		},
		{
			name: "interested",
			message: client.Message{
				ID:      client.MessageInterested,
				Payload: nil,
			},
			expected: []byte{0x0, 0x0, 0x0, 0x1, 0x2},
		},
		{
			name: "not interested",
			message: client.Message{
				ID:      client.MessageNotInterested,
				Payload: nil,
			},
			expected: []byte{0x0, 0x0, 0x0, 0x1, 0x3},
		},
		{
			name: "have",
			message: client.Message{
				ID:      client.MessageHave,
				Payload: []byte{0x1, 0x2, 0x3, 0x4},
			},
			expected: []byte{0x0, 0x0, 0x0, 0x5, 0x4, 0x1, 0x2, 0x3, 0x4},
		},
		{
			name: "bitfield",
			message: client.Message{
				ID:      client.MessageBitfield,
				Payload: []byte{0x1, 0x2, 0x3, 0x4},
			},
			expected: []byte{0x0, 0x0, 0x0, 0x5, 0x5, 0x1, 0x2, 0x3, 0x4},
		},
		{
			name: "request",
			message: client.Message{
				ID:      client.MessageRequest,
				Payload: []byte{0x1, 0x2, 0x3, 0x4},
			},
			expected: []byte{0x0, 0x0, 0x0, 0x5, 0x6, 0x1, 0x2, 0x3, 0x4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.message.Serialize()
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("Serialize() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
