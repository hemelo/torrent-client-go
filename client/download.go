package client

import (
	"Torrent-Client/torrent"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"runtime"
	"time"
)

const Port uint16 = 6881

type PieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type PieceResult struct {
	index int
	data  []byte
}

type PieceProgress struct {
	index      int
	client     *Client
	buffer     []byte
	downloaded int
	requested  int
	backlog    int
}

type DownloadOptions struct {
	path string
}

type DownloadInfo struct {
	peerID       [20]byte
	peers        []Peer
	pieceWork    chan *PieceWork
	pieceResults chan *PieceResult
	torrent      *torrent.TorrentFile
}

func (pw *PieceWork) validate(data []byte) error {

	log.Debug().Int("index", pw.index).Msg("validating piece")

	hash := sha1.Sum(data)

	if hash != pw.hash {
		return fmt.Errorf("hash mismatch for piece %d", pw.index)
	}

	return nil
}

func (pieceProgress *PieceProgress) readMessage() error {
	message, err := pieceProgress.client.ReadMessage()

	if err != nil {
		return err
	}

	if message.ID == MessageKeepAlive {
		return nil
	}

	switch message.ID {
	case MessageChoke:
		pieceProgress.client.choked = true
	case MessageUnchoke:
		pieceProgress.client.choked = false
	case MessageHave:
		index, err := ParseHave(*message)

		if err != nil {
			log.Error().Err(err).Msg("could not parse have message")
		}

		pieceProgress.client.bitfield.SetPiece(index)
	case MessagePiece:
		index, _, err := ParsePiece(pieceProgress.index, pieceProgress.buffer, *message)

		if err != nil {
			log.Error().Err(err).Msg("could not parse piece message")
		}

		pieceProgress.downloaded += index
		pieceProgress.backlog--
	default:
		log.Warn().Int("id", int(message.ID)).Msg("unexpected message")
	}

	return nil
}

func DownloadTorrent(t *torrent.TorrentFile, opts DownloadOptions) error {

	downloadInfo := &DownloadInfo{
		torrent: t,
		peerID:  [20]byte{},
	}

	log.Debug().Str("n ame", t.Name).Msg("starting download for torrent")

	_, err := rand.Read(downloadInfo.peerID[:])

	if err != nil {
		log.Error().Err(err).Msg("failed to generate peer ID")
		return err
	}

	peers, err := t.RequestPeers(downloadInfo.peerID, Port)

	if err != nil {
		log.Error().Err(err).Msg("failed to request peers")
		return err
	}

	downloadInfo.peers = peers

	// Piece work channel is used to send work to workers, each worker will download a piece
	// Results channel is used to send the downloaded piece back to the main thread
	downloadInfo.pieceWork = make(chan *PieceWork, len(t.PiecesHash))
	downloadInfo.pieceResults = make(chan *PieceResult)

	// Create work for each piece
	for index, hash := range t.PiecesHash {
		length := int(t.CalculatePieceSize(index))
		downloadInfo.pieceWork <- &PieceWork{index, hash, length}
	}

	// Start workers to download pieces
	// Each worker will download a piece and send the result back to the main thread
	for _, _ = range peers {
		go startDownloadWorker(downloadInfo)
	}

	// Save data into a buffer till all pieces are downloaded

	buffer := make([]byte, t.Length)
	piecesFinished := 0

	// Wait for all pieces to be downloaded
	for piecesFinished < len(t.PiecesHash) {

		// Get the downloaded piece
		res := <-downloadInfo.pieceResults

		// Calculate the begin and end of the piece in the buffer
		begin, end := t.CalculateBoundsForPiece(res.index)

		// Copy the downloaded piece into the buffer
		copy(buffer[begin:end], res.data)

		piecesFinished++

		percent := float64(piecesFinished) / float64(len(t.PiecesHash)) * 100
		totalWorkers := runtime.NumGoroutine() - 1

		log.Info().Str("name", t.Name).Float64("percent", percent).Int("workers", totalWorkers).Msg("download progress")
	}

	// Close the piece work channel to signal workers to stop
	close(downloadInfo.pieceWork)

	log.Debug().Str("name", t.Name).Msg("creating file to save data from torrent")

	output, err := os.Create(opts.path)

	if err != nil {
		log.Error().Err(err).Str("name", t.Name).Msg("failed to create file")
		return err
	}

	defer func(output *os.File) {
		err := output.Close()
		if err != nil {
			log.Error().Err(err).Str("name", t.Name).Msg("failed to close file")
		} else {
			log.Debug().Str("name", t.Name).Msg("file closed")
		}
	}(output)

	log.Debug().Str("name", t.Name).Msg("writing data to file")

	_, err = output.Write(buffer)

	if err != nil {
		log.Error().Err(err).Str("name", t.Name).Msg("failed to write file")
		return err
	}

	log.Info().Str("name", t.Name).Msg("download completed")

	return nil
}

func startDownloadWorker(dwInfo *DownloadInfo) {

	if len(dwInfo.peers) == 0 {
		log.Error().Msg("no peers available")
		return
	}

	client, err := NewClient(dwInfo.peers[0], dwInfo.torrent.InfoHash, dwInfo.peerID)

	if err != nil {
		log.Error().Err(err).Msg("failed to create client")
		return
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close connection")
		} else {
			log.Debug().Msg("connection closed")
		}
	}(client.conn)

	err = client.SendUnchoke()

	if err != nil {
		log.Error().Err(err).Msg("failed to send unchoke message")
		return
	}

	err = client.SendInterested()

	for pieceWork := range dwInfo.pieceWork {

		// If bitfield does not have the piece, put it back in the queue
		if !client.bitfield.HasPiece(pieceWork.index) {
			dwInfo.pieceWork <- pieceWork
			continue
		}

		buffer, err := downloadPiece(client, pieceWork)

		// If fails to download piece, put it back in the queue
		if err != nil {
			log.Error().Err(err).Msg("failed to download piece")
			dwInfo.pieceWork <- pieceWork
			continue
		}

		// Validate the piece, if it fails, put it back in the queue
		if err := pieceWork.validate(buffer); err != nil {
			log.Error().Err(err).Msg("failed to validate piece")
			dwInfo.pieceWork <- pieceWork
			continue
		}

		err = client.SendHave(pieceWork.index)

		if err != nil {
			log.Error().Err(err).Msg("failed to send have message")
		}

		dwInfo.pieceResults <- &PieceResult{pieceWork.index, buffer}
	}
}

func downloadPiece(client *Client, pieceWork *PieceWork) ([]byte, error) {

	pieceProgress := PieceProgress{
		index:  pieceWork.index,
		client: client,
		buffer: make([]byte, pieceWork.length),
	}

	err := client.conn.SetDeadline(time.Now().Add(defaultPeerTimeout))

	if err != nil {
		log.Error().Err(err).Msg("failed to set deadline")
		return nil, err
	}

	defer func(conn net.Conn, t time.Time) {
		err := conn.SetDeadline(t)
		if err != nil {
			log.Error().Err(err).Msg("failed to reset deadline")
		}
	}(client.conn, time.Time{})

	for pieceProgress.downloaded < pieceWork.length {

		if pieceProgress.client.choked {

			err := pieceProgress.readMessage()

			if err != nil {
				log.Error().Err(err).Msg("failed to read message")
				return nil, err
			}

			continue
		}

		if pieceProgress.backlog < maxRequests && pieceProgress.requested < pieceWork.length {

			blockSize := maxBlockSize

			if pieceWork.length-pieceProgress.requested < blockSize {
				blockSize = pieceWork.length - pieceProgress.requested
			}

			err := client.SendRequest(pieceWork.index, pieceProgress.requested, blockSize)

			if err != nil {
				log.Error().Err(err).Msg("failed to send request")
				return nil, err
			}

			pieceProgress.requested += blockSize
			pieceProgress.backlog++
		}

		err := pieceProgress.readMessage()

		if err != nil {
			log.Error().Err(err).Msg("failed to read message")
			return nil, err
		}
	}

	return pieceProgress.buffer, nil
}
