package torrent

import (
	"Torrent-Client/bencode"
	"Torrent-Client/peers"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

type TrackerResponse struct {
	Interval int
	Peers    string
}

type BencodeToTrackerResponseOpts struct {
	from string
}

// Builds the tracker URL for the torrent file
func (t *TorrentFile) buildTrackerUrl(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)

	if err != nil {
		log.Error().Err(err).Str("announce", t.Announce).Msg("failed to parse announce URL")
		return "", err
	}

	params := url.Values{

		// The tracker will use this to figure out which peers to show us
		// The info hash is a SHA1 hash of the bencoded info key from the metainfo file
		"info_hash": []string{string(t.InfoHash[:])},

		// The peer ID is a 20-byte string used as a unique ID for the client
		// This is used to identify the client to the tracker
		"peer_id": []string{string(peerID[:])},

		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(int(t.Length))},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

// Peers is a list of peers that the client can connect to
// First, its required to send an HTTP GET request to the tracker
// The tracker will respond with a bencoded dictionary
// The dictionary will contain a list of peers
// Each peer is a dictionary containing the IP address and port number
// It's made of 6 bytes for the IP address and 2 bytes for the port number
// Big-endian notation is used for both the IP address and port number
func (t *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {

	url, err := t.buildTrackerUrl(peerID, port)

	if err != nil {
		log.Error().Err(err).Msg("could not build tracker URL to request peers")
		return nil, err
	}

	c := &http.Client{Timeout: defaultHTTPTimeout}

	resp, err := c.Get(url)

	if err != nil {
		log.Error().Err(err).Msg("failed to send GET request to tracker")
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close response body")
		}
	}(resp.Body)

	result, err := bencode.Parse(resp.Body)

	if err != nil {
		log.Error().Err(err).Msg("failed to parse tracker response")
		return nil, err
	}

	trackerResponse, err := BencodeToTrackerResponse(result, BencodeToTrackerResponseOpts{from: url})

	if err != nil {
		log.Error().Err(err).Msg("failed to convert tracker response")
		return nil, err
	}

	return peers.DecodePeers([]byte(trackerResponse.Peers))
}

func BencodeToTrackerResponse(result bencode.BencodeValue, opts BencodeToTrackerResponseOpts) (TrackerResponse, error) {

	// Check the parsed result
	if result.Type != bencode.DictType {
		log.Error().Str("from", opts.from).Msg("expected DictType")
		return TrackerResponse{}, fmt.Errorf("expected DictType, got %v", result.Type)
	}

	// Check the interval field
	interval, ok := result.Dict["interval"]

	if !ok {
		log.Error().Str("from", opts.from).Msg("missing interval")
		return TrackerResponse{}, fmt.Errorf("missing interval")
	} else if interval.Type != bencode.IntegerType {
		log.Error().Str("from", opts.from).Msg("interval is not an integer")
		return TrackerResponse{}, fmt.Errorf("interval is not an integer")
	} else {
		log.Debug().Str("from", opts.from).Int64("interval", interval.Int).Msg("interval")
	}

	// Check the peers field
	peers, ok := result.Dict["peers"]

	if !ok {
		log.Error().Str("from", opts.from).Msg("missing peers")
		return TrackerResponse{}, fmt.Errorf("missing peers")
	} else if peers.Type != bencode.StringType {
		log.Error().Str("from", opts.from).Msg("peers is not a string")
		return TrackerResponse{}, fmt.Errorf("peers is not a string")
	} else {
		log.Debug().Str("from", opts.from).Str("peers", peers.Str).Msg("peers")
	}

	return TrackerResponse{
		Interval: int(interval.Int),
		Peers:    peers.Str,
	}, nil
}
