package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

const peerSize = 6

type Peer struct {
	IP   net.IP
	Port uint16
}

func DecodePeers(bytes []byte) ([]Peer, error) {

	if len(bytes) == 0 {
		return nil, fmt.Errorf("received empty peers")
	}

	numPeers := len(bytes) / peerSize

	if len(bytes)%peerSize != 0 {
		return nil, fmt.Errorf("received malformed peers")
	}

	peers := make([]Peer, numPeers)

	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i] = Peer{
			IP:   net.IP(bytes[offset : offset+4]),
			Port: binary.BigEndian.Uint16(bytes[offset+4 : offset+6]),
		}
	}

	return peers, nil
}

func (p Peer) Address() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

func (p Peer) String() string {
	return p.Address()
}
