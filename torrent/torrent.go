package torrent

import (
	"Torrent-Client/bencode"
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"
)

type TorrentFile struct {
	Name         string
	Announce     string
	Comment      string
	CreatedBy    string
	CreationDate int64
	InfoHash     [20]byte
	PiecesHash   [][20]byte
	PieceLength  int64
	Length       int64
}

type BencodeToTorrentFileOpts struct {
	from string
}

func NewTorrentFrom(path string) (TorrentFile, error) {

	log.Debug().Str("from", path).Msg("reading torrent file")

	data, err := os.ReadFile(path)

	if err != nil {
		log.Error().Err(err).Str("from", path).Msg("failed to read file")
		return TorrentFile{}, err
	}

	reader := io.Reader(bufio.NewReader(bytes.NewReader(data)))

	log.Debug().Str("from", path).Msg("parsing torrent file")

	result, err := bencode.Parse(reader)

	if err != nil {
		log.Error().Err(err).Str("from", path).Msg("failed to parse torrent file")
		return TorrentFile{}, err
	}

	return BencodeToTorrentFile(result, BencodeToTorrentFileOpts{from: path})
}

func BencodeToTorrentFile(result bencode.BencodeValue, opts BencodeToTorrentFileOpts) (TorrentFile, error) {
	// Check the parsed result
	if result.Type != bencode.DictType {
		log.Error().Str("from", opts.from).Msg("expected DictType")
		return TorrentFile{}, fmt.Errorf("expected DictType, got %v", result.Type)
	}

	// Check the announce URL
	announce, ok := result.Dict["announce"]

	if !ok {
		log.Error().Str("from", opts.from).Msg("missing announce URL")
		return TorrentFile{}, fmt.Errorf("missing announce URL")
	} else if announce.Type != bencode.StringType {
		log.Error().Str("from", opts.from).Msg("announce URL is not a string")
		return TorrentFile{}, fmt.Errorf("announce URL is not a string")
	} else {
		log.Debug().Str("from", opts.from).Str("announce", announce.Str).Msg("announce URL")
	}

	creationDate, ok := result.Dict["creation date"]

	if !ok {
		log.Debug().Str("from", opts.from).Msg("missing creation date")
	} else if creationDate.Type != bencode.IntegerType {
		log.Error().Str("from", opts.from).Msg("creation date is not an integer")
		return TorrentFile{}, fmt.Errorf("creation date is not an integer")
	} else {
		log.Debug().Str("from", opts.from).Int64("creation date", creationDate.Int).Msg("creation date")
	}

	createdBy, ok := result.Dict["created by"]

	if !ok {
		log.Debug().Str("from", opts.from).Msg("missing created by")
	} else if createdBy.Type != bencode.StringType {
		log.Error().Str("from", opts.from).Msg("created by is not a string")
		return TorrentFile{}, fmt.Errorf("created by is not a string")
	} else {
		log.Debug().Str("from", opts.from).Str("created by", createdBy.Str).Msg("created by")
	}

	comment, ok := result.Dict["comment"]

	if !ok {
		log.Debug().Str("from", opts.from).Msg("missing comment")
	} else if comment.Type != bencode.StringType {
		log.Error().Str("from", opts.from).Msg("comment is not a string")
		return TorrentFile{}, fmt.Errorf("comment is not a string")
	} else {
		log.Debug().Str("from", opts.from).Str("comment", comment.Str).Msg("comment")
	}

	//...

	// Check the info field
	info, ok := result.Dict["info"]

	if !ok {
		log.Error().Str("from", opts.from).Msg("missing info")
		return TorrentFile{}, fmt.Errorf("missing info")
	} else if info.Type != bencode.DictType {
		log.Error().Str("from", opts.from).Msg("info is not a dictionary")
		return TorrentFile{}, fmt.Errorf("info is not a dictionary")
	} else {
		log.Debug().Str("from", opts.from).Msg("info dictionary")
	}

	// Check the name field in the info dictionary
	name, ok := info.Dict["name"]

	if !ok {
		log.Error().Str("from", opts.from).Msg("missing name")
		return TorrentFile{}, fmt.Errorf("missing name")
	} else if name.Type != bencode.StringType {
		log.Error().Str("from", opts.from).Msg("name is not a string")
		return TorrentFile{}, fmt.Errorf("name is not a string")
	} else {
		log.Debug().Str("from", opts.from).Str("name", name.Str).Msg("name")
	}

	length, ok := info.Dict["length"]

	if !ok {
		log.Error().Str("from", opts.from).Msg("missing length")
	} else if length.Type != bencode.IntegerType {
		log.Error().Str("from", opts.from).Msg("length is not an integer")
		return TorrentFile{}, fmt.Errorf("length is not an integer")
	} else {
		log.Debug().Str("from", opts.from).Int64("length", length.Int).Msg("length")
	}

	// Check the piece length
	pieceLength, ok := info.Dict["piece length"]

	if !ok {
		log.Error().Str("from", opts.from).Msg("missing piece length")
		return TorrentFile{}, fmt.Errorf("missing piece length")
	} else if pieceLength.Type != bencode.IntegerType {
		log.Error().Str("from", opts.from).Msg("piece length is not an integer")
		return TorrentFile{}, fmt.Errorf("piece length is not an integer")
	} else {
		log.Debug().Str("from", opts.from).Int64("piece length", pieceLength.Int).Msg("piece length")
	}

	// Check the pieces field
	pieces, ok := info.Dict["pieces"]

	if !ok {
		log.Error().Str("from", opts.from).Msg("missing pieces")
	} else if pieces.Type != bencode.StringType {
		log.Error().Str("from", opts.from).Msg("pieces is not a string")
		return TorrentFile{}, fmt.Errorf("pieces is not a string")
	} else {
		log.Debug().Str("from", opts.from).Msg("pieces")
	}

	piecesHashes, err := splitPiecesInHashes(pieces)

	if err != nil {
		log.Error().Err(err).Str("from", opts.from).Msg("failed to split pieces in hashes")
		return TorrentFile{}, err
	}

	infoHash, err := hashInfo(info)

	if err != nil {
		log.Error().Err(err).Str("from", opts.from).Msg("failed to hash info")
		return TorrentFile{}, err
	}

	torrent := TorrentFile{
		Name:         name.Str,
		Announce:     announce.Str,
		Comment:      comment.Str,
		CreatedBy:    createdBy.Str,
		CreationDate: creationDate.Int,
		PieceLength:  pieceLength.Int,
		Length:       length.Int,
		PiecesHash:   piecesHashes,
		InfoHash:     infoHash,
	}

	return torrent, nil
}

func hashInfo(info bencode.BencodeValue) ([20]byte, error) {

	buffer := bytes.Buffer{}

	err := info.Encode(&buffer)

	if err != nil {
		log.Error().Err(err).Msg("failed to encode info")
		return [20]byte{}, err
	}

	hash := sha1.Sum(buffer.Bytes())

	return hash, nil
}

func splitPiecesInHashes(pieces bencode.BencodeValue) ([][20]byte, error) {
	hashLength := 20
	buffer := []byte(pieces.Str)

	if len(buffer)%hashLength != 0 {
		log.Error().Int("length", len(buffer)).Msg("invalid pieces length")
		return nil, fmt.Errorf("invalid pieces length")
	}

	totalHashes := len(buffer) / hashLength
	hashes := make([][20]byte, totalHashes)

	for i := 0; i < totalHashes; i++ {
		copy(hashes[i][:], buffer[i*hashLength:(i+1)*hashLength])
	}

	return hashes, nil
}
