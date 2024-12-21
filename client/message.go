package client

import (
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
)

type MessageID int8 // Message ID is a single byte

// States
const (
	MessageKeepAlive     MessageID = iota - 1 // KeepAlive is a message that tells the peer that the client is still connected
	MessageChoke                              // Choke is a message that tells the peer to stop sending requests
	MessageUnchoke                            // Unchoke is a message that tells the peer it can send requests
	MessageInterested                         // Interested is a message that tells the peer that the client is interested in a piece
	MessageNotInterested                      // NotInterested is a message that tells the peer that the client is not interested in a piece
	MessageHave                               // Have is a message that tells the peer that the client has a piece
	MessageBitfield                           // Bitfield is a message that tells the peer which pieces the client has
	MessageRequest                            // Request is a message that tells the peer that the client wants a piece
	MessagePiece                              // Piece is a message that contains a piece
	MessageCancel                             // Cancel is a message that tells the peer that the client no longer wants a piece
)

type Message struct {
	ID      MessageID
	Payload []byte
}

func NewRequestMessage(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return &Message{
		ID:      MessageRequest,
		Payload: payload,
	}
}

func NewHaveMessage(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))

	return &Message{
		ID:      MessageHave,
		Payload: payload,
	}
}

func NewInterestedMessage() *Message {
	return &Message{
		ID:      MessageInterested,
		Payload: nil,
	}
}

func NewNotInterestedMessage() *Message {
	return &Message{
		ID:      MessageNotInterested,
		Payload: nil,
	}
}

func NewUnchokeMessage() *Message {
	return &Message{
		ID:      MessageUnchoke,
		Payload: nil,
	}
}

func ReadBitfieldMessage(reader io.Reader) (Bitfield, error) {

	log.Debug().Msg("reading bitfield message")

	msg, err := ReadMessage(reader)

	if err != nil {
		log.Error().Err(err).Msg("failed to read message")
		return nil, err
	}

	if msg.ID != MessageBitfield {
		log.Error().Msg("expected bitfield message")
		return nil, fmt.Errorf("expected bitfield message, got %v", msg.ID)
	}

	return msg.Payload, nil
}

func ReadMessage(reader io.Reader) (*Message, error) {

	log.Debug().Msg("reading message")

	if reader == nil {
		log.Error().Msg("reader is nil")
		return nil, fmt.Errorf("reader is nil")
	}

	lengthBuffer := make([]byte, 4)

	_, err := io.ReadFull(reader, lengthBuffer)

	if err != nil {
		log.Error().Err(err).Msg("failed to read message length")
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuffer)

	if length == 0 {
		log.Debug().Msg("message length is 0, keep-alive message")
		return &Message{
			ID: MessageKeepAlive,
		}, nil
	}

	messageBuffer := make([]byte, length)

	_, err = io.ReadFull(reader, messageBuffer)

	if err != nil {
		log.Error().Err(err).Msg("failed to read message")
		return nil, err
	}

	data := &Message{
		ID:      MessageID(messageBuffer[0]),
		Payload: messageBuffer[1:],
	}

	log.Debug().Int("id", int(data.ID)).Bytes("payload", data.Payload).Str("summary", data.Type()).Msg("read message")
	return data, nil
}

func (m *Message) Serialize() []byte {

	if m.ID == MessageKeepAlive {
		return make([]byte, 4)
	}

	length := 1 + len(m.Payload)
	buffer := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buffer[0:4], uint32(length))
	buffer[4] = byte(m.ID)
	copy(buffer[5:], m.Payload)

	return buffer
}

func (m *Message) String() string {
	return fmt.Sprintf("%s: %v", m.Type(), m.Payload)
}

func (m *Message) Type() string {
	switch m.ID {
	case MessageKeepAlive:
		return "keep alive"
	case MessageChoke:
		return "choke"
	case MessageUnchoke:
		return "unchoke"
	case MessageInterested:
		return "interested"
	case MessageNotInterested:
		return "not interested"
	case MessageHave:
		return "have"
	case MessageBitfield:
		return "bitfield"
	case MessageRequest:
		return "request"
	case MessagePiece:
		return "piece"
	case MessageCancel:
		return "cancel"
	default:
		return "unknown"
	}
}

func ParseHave(message Message) (int, error) {

	if message.ID != MessageHave {
		log.Error().Int("id", int(message.ID)).Int("expected", int(MessageHave)).Msg("unexpected message")
		return 0, fmt.Errorf("unexpected message")
	}

	if len(message.Payload) != 4 {
		log.Error().Int("length", len(message.Payload)).Int("expected", 4).Msg("unexpected payload length")
		return 0, fmt.Errorf("unexpected payload length")
	}

	index := int(binary.BigEndian.Uint32(message.Payload))
	return index, nil
}

func ParsePiece(index int, buffer []byte, message Message) (int, []byte, error) {

	if message.ID != MessagePiece {
		log.Error().Int("id", int(message.ID)).Int("expected", int(MessagePiece)).Msg("unexpected message")
		return 0, nil, fmt.Errorf("unexpected message")
	}

	if len(message.Payload) < 8 {
		log.Error().Int("length", len(message.Payload)).Int("expected", 8).Msg("unexpected payload length")
		return 0, nil, fmt.Errorf("unexpected payload length")
	}

	indexToCheck := int(binary.BigEndian.Uint32(message.Payload[0:4]))

	if indexToCheck != index {
		log.Error().Int("parsed", indexToCheck).Int("expected", index).Msg("unexpected piece index")
		return 0, nil, fmt.Errorf("unexpected piece index %d", indexToCheck)
	}

	begin := int(binary.BigEndian.Uint32(message.Payload[4:8]))

	if begin < 0 {
		log.Error().Int("begin", begin).Msg("unexpected begin")
		return 0, nil, fmt.Errorf("unexpected begin %d", begin)
	}

	if begin >= len(buffer) {
		log.Error().Int("begin", begin).Int("length", len(buffer)).Msg("begin exceeds buffer length")
		return 0, nil, fmt.Errorf("begin exceeds buffer length")
	}

	data := message.Payload[8:]

	if begin+len(data) > len(buffer) {
		log.Error().Int("begin", begin).Int("length", len(data)).Int("buffer", len(buffer)).Msg("data exceeds buffer length")
		return 0, nil, fmt.Errorf("data exceeds buffer length")
	}

	copy(buffer[begin:], data)
	return len(data), data, nil
}
