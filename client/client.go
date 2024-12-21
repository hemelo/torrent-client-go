package client

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

const maxBlockSize = 16384
const maxRequests = 5
const defaultPeerTimeout = 3 * time.Second

type Client struct {
	name       string
	conn       net.Conn
	choked     bool
	interested bool
	bitfield   Bitfield
	peer       Peer
	infoHash   [20]byte
	peerID     [20]byte
}

func NewClient(peer Peer, infoHash [20]byte, peerID [20]byte) (*Client, error) {

	log.Debug().Str("peer", peer.Address()).Msg("connecting to peer")

	// Dial is a function that connects to the address on the named network
	// The network must be "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "ip", "ip4", "ip6"
	conn, err := net.DialTimeout("tcp", peer.Address(), defaultPeerTimeout)

	if err != nil {
		log.Debug().Err(err).Str("peer", peer.Address()).Msg("failed to connect to peer")
		return nil, err
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error().Err(err).Str("peer", peer.Address()).Msg("failed to close connection")
		}
	}(conn)

	log.Debug().Str("peer", peer.Address()).Msg("handshaking with peer")

	err = handshake(conn, peer, infoHash, peerID)

	if err != nil {
		log.Debug().Err(err).Str("peer", peer.Address()).Msg("failed to do handshake")
		return nil, err
	}

	log.Debug().Str("peer", peer.Address()).Msg("handshake successful")

	bitfield, err := ReadBitfieldMessage(conn)

	if err != nil {
		log.Error().Err(err).Str("peer", peer.Address()).Msg("failed to read bitfield message")
		return nil, err
	}

	return &Client{
		peer:     peer,
		infoHash: infoHash,
		bitfield: bitfield,
		peerID:   peerID,
		conn:     conn,
		choked:   true,
	}, nil
}

func (c *Client) ReadMessage() (*Message, error) {

	if c.conn == nil {
		log.Error().Str("peer", c.peer.Address()).Msg("connection is nil")
		return nil, fmt.Errorf("connection is nil")
	}

	return ReadMessage(c.conn)
}

func (c *Client) SendRequest(index, begin, length int) error {
	message := NewRequestMessage(index, begin, length)
	_, err := c.conn.Write(message.Serialize())

	if err != nil {
		log.Error().Err(err).Str("peer", c.peer.Address()).Str("message", message.String()).Msg("failed to write message")
	}

	return err
}

func (c *Client) SendHave(index int) error {
	message := NewHaveMessage(index)
	_, err := c.conn.Write(message.Serialize())

	if err != nil {
		log.Error().Err(err).Str("peer", c.peer.Address()).Str("message", message.String()).Msg("failed to write message")
	}

	return err
}

func (c *Client) SendInterested() error {
	message := NewInterestedMessage()
	_, err := c.conn.Write(message.Serialize())

	if err != nil {
		log.Error().Err(err).Str("peer", c.peer.Address()).Str("message", message.String()).Msg("failed to write message")
	}

	return err
}

func (c *Client) SendNotInterested() error {
	message := NewNotInterestedMessage()
	_, err := c.conn.Write(message.Serialize())

	if err != nil {
		log.Error().Err(err).Str("peer", c.peer.Address()).Str("message", message.String()).Msg("failed to write message")
	}

	return err
}

func (c *Client) SendUnchoke() error {
	message := NewUnchokeMessage()
	_, err := c.conn.Write(message.Serialize())

	if err != nil {
		log.Error().Err(err).Str("peer", c.peer.Address()).Str("message", message.String()).Msg("failed to write message")
	}

	return err
}
