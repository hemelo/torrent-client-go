package client

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"time"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(peerID [20]byte, infoHash [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func handshake(conn net.Conn, peer Peer, infoHash [20]byte, peerID [20]byte) error {

	log.Debug().Str("peer", peer.Address()).Msg("setting deadline")

	err := conn.SetDeadline(time.Now().Add(defaultPeerTimeout))

	if err != nil {
		log.Error().Err(err).Str("peer", peer.Address()).Msg("failed to set deadline")
		return err
	}

	defer func(conn net.Conn, t time.Time) {

		log.Debug().Str("peer", peer.Address()).Msg("resetting deadline")

		err := conn.SetDeadline(t)
		if err != nil {
			log.Error().Err(err).Str("peer", peer.Address()).Msg("failed to reset deadline")
		}
	}(conn, time.Time{})

	handshakeRequest := NewHandshake(peerID, infoHash)

	log.Debug().Str("peer", peer.Address()).Msg("writing handshake")

	_, err = conn.Write(handshakeRequest.Serialize())

	if err != nil {
		log.Error().Err(err).Str("peer", peer.Address()).Msg("failed to write handshake")
		return err
	}

	handshakeResponse, err := ReadResponse(conn)

	if err != nil {
		log.Error().Err(err).Str("peer", peer.Address()).Msg("failed to read handshake")
		return err
	}

	log.Debug().Str("peer", peer.Address()).Msg("read handshake with success, checking info hash")

	if !bytes.Equal(handshakeResponse.InfoHash[:], infoHash[:]) {
		log.Error().Str("peer", peer.Address()).Str("InfoHash", string(handshakeResponse.InfoHash[:])).Str("expected", string(infoHash[:])).Msg("info hash mismatch")
		return fmt.Errorf("info hash mismatch")
	}

	return nil
}

func ReadResponse(r io.Reader) (*Handshake, error) {
	buffer := make([]byte, 1)

	_, err := io.ReadFull(r, buffer)

	if err != nil {
		log.Debug().Err(err).Msg("failed to read handshake")
		return nil, err
	}

	protocolStringLen := int(buffer[0])

	if protocolStringLen == 0 {
		log.Debug().Msg("protocol string length is 0")
		return nil, fmt.Errorf("protocol string length is 0")
	}

	handshakeBuffer := make([]byte, protocolStringLen+48)

	_, err = io.ReadFull(r, handshakeBuffer)

	if err != nil {
		log.Debug().Err(err).Msg("failed to read handshake")
		return nil, err
	}

	return &Handshake{
		Pstr:     string(handshakeBuffer[0:protocolStringLen]),
		InfoHash: [20]byte(handshakeBuffer[protocolStringLen+8 : protocolStringLen+20+8]),
		PeerID:   [20]byte(handshakeBuffer[protocolStringLen+20+8 : protocolStringLen+20+8+20]),
	}, nil
}

func (h *Handshake) Serialize() []byte {

	log.Debug().Msg("serializing handshake")

	buffer := make([]byte, len(h.Pstr)+20+20+8+1)
	curr := copy(buffer[0:], string(rune(len(h.Pstr))))
	curr += copy(buffer[1:], h.Pstr)
	curr += copy(buffer[curr:], make([]byte, 8))
	curr += copy(buffer[curr:], h.InfoHash[:])
	curr += copy(buffer[curr:], h.PeerID[:])
	return buffer
}
