package bencode

import (
	"bufio"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"strconv"
	"sync"
)

type BencodeValue struct {
	Type  BencodeType
	Uint  uint64
	Int   int64
	Float float64
	Str   string
	List  []BencodeValue
	Dict  map[string]BencodeValue
}

type BencodeType int

const (
	IntegerType BencodeType = iota
	UnsignedIntegerType
	FloatType
	StringType
	ListType
	DictType
)

func parseInteger(reader *bufio.Reader) (BencodeValue, error) {

	buffer, err := readUntil(reader, 'e')

	if err != nil {
		return BencodeValue{}, err
	}

	content := string(buffer)

	if _int64, err := strconv.ParseInt(content, 10, 64); err == nil {
		return BencodeValue{Type: IntegerType, Int: _int64}, nil
	}

	if _uint64, err := strconv.ParseUint(content, 10, 64); err == nil {
		return BencodeValue{Type: UnsignedIntegerType, Uint: _uint64}, nil
	}

	if _float64, err := strconv.ParseFloat(content, 64); err == nil {
		return BencodeValue{Type: FloatType, Float: _float64}, nil
	}

	return BencodeValue{}, fmt.Errorf("invalid integer")
}

func parseString(reader *bufio.Reader) (BencodeValue, error) {

	err := reader.UnreadByte()

	if err != nil {
		return BencodeValue{}, err
	}

	str, err := decodeString(reader)

	if err != nil {
		return BencodeValue{}, err
	}

	return BencodeValue{Type: StringType, Str: str}, nil
}

func parseList(reader *bufio.Reader) (BencodeValue, error) {

	list := make([]BencodeValue, 0)

	for {
		c, err := reader.ReadByte()

		if err != nil {
			return BencodeValue{}, err
		}

		if c == 'e' {
			break
		}

		err = reader.UnreadByte()

		if err != nil {
			return BencodeValue{}, err
		}

		value, err := parse(reader)

		if err != nil {
			return BencodeValue{}, err
		}

		list = append(list, value)
	}

	return BencodeValue{Type: ListType, List: list}, nil
}

func parseDict(reader *bufio.Reader) (BencodeValue, error) {

	dict := make(map[string]BencodeValue)

	for {
		c, err := reader.ReadByte()

		if err != nil {
			return BencodeValue{}, err
		}

		if c == 'e' {
			break
		}

		err = reader.UnreadByte()

		if err != nil {
			return BencodeValue{}, err
		}

		key, err := decodeString(reader)

		if err != nil {
			return BencodeValue{}, err
		}

		value, err := parse(reader)

		if err != nil {
			return BencodeValue{}, err
		}

		dict[key] = value
	}

	return BencodeValue{Type: DictType, Dict: dict}, nil
}

func parse(reader *bufio.Reader) (BencodeValue, error) {

	c, bErr := reader.ReadByte()

	if bErr != nil {
		return BencodeValue{}, bErr
	}

	var result BencodeValue
	var err error

	log.Debug().Str("character", string(c)).Msg("parsing character")

	switch c {
	case 'i':
		result, err = parseInteger(reader)
	case 'l':
		result, err = parseList(reader)
	case 'd':
		result, err = parseDict(reader)
	default:
		if c >= '0' && c <= '9' {
			result, err = parseString(reader)
		} else {
			result, err = BencodeValue{}, fmt.Errorf("invalid character")
		}
	}

	return result, err
}

func Parse(reader io.Reader) (BencodeValue, error) {

	if bReader, ok := reader.(*bufio.Reader); ok {
		return parse(bReader)
	}

	/*
		The code below initializes a `sync.Pool` to manage a pool of `bufio.Reader` objects. It retrieves a `bufio.Reader`
		from the pool if available, or creates a new one if not. The `bufio.Reader` is reset with the new `reader` and returned
		to the pool after use. This approach optimizes performance by reusing `bufio.Reader` instances, reducing the overhead
		of repeated allocations.
	*/
	bufioReaderPool := sync.Pool{}

	bReader := func() *bufio.Reader {

		if v := bufioReaderPool.Get(); v != nil {
			br := v.(*bufio.Reader)
			br.Reset(reader)
			return br
		}

		return bufio.NewReader(reader)
	}()

	defer bufioReaderPool.Put(bReader)

	return parse(bReader)
}

// Read until the delimiter byte is found in the reader.
func readUntil(reader *bufio.Reader, delim byte) ([]byte, error) {
	data, err := reader.ReadSlice(delim)

	if err != nil {
		return nil, err
	}

	lenData := len(data)

	if lenData > 0 {
		data = data[:lenData-1]
	} else {
		panic("missed read")
	}

	return data, nil
}

// Read exactly len(buffer) bytes from the reader into the buffer.
func readFull(reader *bufio.Reader, buffer []byte) (int, error) {
	return readAtLeast(reader, buffer, len(buffer))
}

// Read at least min bytes from the reader into the buffer.
func readAtLeast(reader *bufio.Reader, buffer []byte, min int) (n int, err error) {
	if len(buffer) < min {
		return 0, fmt.Errorf("buffer too small")
	}

	// Read at least min bytes.
	for n < min && err == nil {
		var nn int
		nn, err = reader.Read(buffer[n:])
		n += nn
	}

	if n == min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}

	return
}

func decodeString(reader *bufio.Reader) (string, error) {

	log.Debug().Msg("getting length of content before ':' character")

	length, err := decodeInt64(reader, ':')

	if err != nil {
		log.Error().Err(err).Msg("could not decode content length")
		return "", err
	}

	if length < 0 {
		log.Debug().Int64("length", length).Msg("invalid content length")
		return "", fmt.Errorf("invalid string length")
	} else {
		log.Debug().Int64("length", length).Msg("got length of content before ':' character")
	}

	if peekBuffer, peekErr := reader.Peek(int(length)); peekErr == nil {
		data := string(peekBuffer)
		_, err := reader.Discard(int(length))

		log.Debug().Str("data", data).Msg("peeked content")

		if err != nil {
			log.Error().Err(err).Msg("could not discard content")
		}

		return data, err
	}

	log.Error().Msg("could not peek content, will try to read full content")

	buffer := make([]byte, length)

	if _, err := readFull(reader, buffer); err != nil {
		log.Error().Err(err).Msg("could not read content")
		return "", err
	}

	data := string(buffer)
	log.Debug().Str("data", data).Msg("attempted to read content successfully")
	return data, nil
}

func decodeInt64(reader *bufio.Reader, delim byte) (int64, error) {

	log.Debug()

	buffer, err := readUntil(reader, delim)

	if err != nil {
		return 0, err
	}

	content := string(buffer)

	if _int64, err := strconv.ParseInt(content, 10, 64); err == nil {
		log.Debug().Int64("int64", _int64).Msg("parsed integer")
		return _int64, nil
	}

	log.Error().Err(err).Msg("invalid integer")
	return 0, fmt.Errorf("invalid integer")
}
