package rdb

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Rdb struct {
	header   RdbHeader
	metadata map[string]string
	Database map[int]map[string]utils.Entry
	checksum string
}

type RdbHeader struct {
	magic   string
	version string
}

const (
	MAGIC_WORD                   string = "REDIS"
	MAGIC_LENGTH                 int    = 5
	VERSION_LENGTH               int    = 4
	METADATA_OPCODE              byte   = 0xFA
	HASH_TABLE_OPCODE            byte   = 0xFB
	EXPIRY_MILLISECONDS          byte   = 0xFC
	EXPIRY_MILLISECONDS_BYTE_LEN int    = 8
	EXPIRY_SECONDS               byte   = 0xFD
	EXPIRY_SECONDS_BYTE_LEN      int    = 4
	DATABASE_OPCODE              byte   = 0xFE
	CHECKSUM_OPCODE              byte   = 0xFF
	CHECKSUM_LENGTH              int    = 8
)

func NewRdbFromFile(dir string, filename string) (*Rdb, error) {
	file, err := os.Open(filepath.Join(dir, filename))
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	return newRdbFromReader(reader)

}

func newRdbFromReader(reader *bufio.Reader) (*Rdb, error) {
	header, err := getHeader(reader)
	if err != nil {
		return nil, err
	}
	metadata, err := getMetadata(reader)
	if err != nil {
		return nil, err
	}

	database, err := getDatabase(reader)
	if err != nil {
		return nil, err
	}
	checksum, err := getChecksum(reader)
	if err != nil {
		return nil, err
	}
	return &Rdb{
		header:   *header,
		metadata: metadata,
		Database: database,
		checksum: checksum,
	}, nil
}

func getHeader(reader *bufio.Reader) (*RdbHeader, error) {
	magic := ""
	for range MAGIC_LENGTH {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		magic += string(b)
	}
	if magic != MAGIC_WORD {
		return nil, fmt.Errorf("expected %s, got %s", MAGIC_WORD, magic)
	}
	version := ""
	for range VERSION_LENGTH {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		version += string(b)
	}
	return &RdbHeader{magic: magic, version: version}, nil
}

func getMetadata(reader *bufio.Reader) (map[string]string, error) {
	metadata := make(map[string]string)
	for {
		next, err := reader.Peek(1)
		if err != nil {
			return nil, err
		}
		if next[0] != METADATA_OPCODE {
			break
		}
		_, err = reader.ReadByte()
		if err != nil {
			return nil, err
		}
		name, err := getStringFromStringEncoding(reader)
		if err != nil {
			return nil, err
		}
		value, err := getStringFromStringEncoding(reader)
		if err != nil {
			return nil, err
		}
		metadata[name] = value
	}
	return metadata, nil
}

func getDatabase(reader *bufio.Reader) (map[int]map[string]utils.Entry, error) {
	database := make(map[int]map[string]utils.Entry)
	for {
		bs, err := reader.Peek(1)
		if err != nil {
			return nil, err
		}
		if bs[0] != DATABASE_OPCODE {
			break
		}
		idx, databaseSection, err := getDatabaseSection(reader)
		if err != nil {
			return nil, err
		}
		database[idx] = databaseSection
	}
	return database, nil
}

func getDatabaseSection(reader *bufio.Reader) (int, map[string]utils.Entry, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return -1, nil, err
	}
	if b != DATABASE_OPCODE {
		return -1, nil, fmt.Errorf("expected %b, got %b", DATABASE_OPCODE, b)
	}
	dbIdx, err := getLengthFromStringEncoding(reader)
	if err != nil {
		return -1, nil, err
	}
	b, err = reader.ReadByte()
	if err != nil {
		return -1, nil, err
	}
	if b != HASH_TABLE_OPCODE {
		return -1, nil, fmt.Errorf("expected %q, got %q,", HASH_TABLE_OPCODE, b)
	}
	for range 2 {
		_, err := getLengthFromStringEncoding(reader)
		if err != nil {
			return -1, nil, err
		}
	}
	databaseSection := make(map[string]utils.Entry)
	for {
		next, err := reader.Peek(1)
		if err != nil {
			return -1, nil, err
		}
		if next[0] == DATABASE_OPCODE || next[0] == CHECKSUM_OPCODE {
			break
		}
		key, entry, err := getEntry(reader)
		if err != nil {
			return -1, nil, err
		}
		databaseSection[key] = *entry
	}
	return dbIdx, databaseSection, nil
}

func getEntry(reader *bufio.Reader) (string, *utils.Entry, error) {
	var expiryTime time.Time
	next, err := reader.Peek(1)
	if err != nil {
		return "", nil, err
	}
	if next[0] == EXPIRY_MILLISECONDS || next[0] == EXPIRY_SECONDS {
		expiryTime, err = getExpiryTime(reader)
		if err != nil {
			return "", nil, err
		}
	}
	_, err = reader.ReadByte() // Assume just string entry for now
	if err != nil {
		return "", nil, err
	}
	key, err := getStringFromStringEncoding(reader)
	if err != nil {
		return "", nil, err
	}
	val, err := getStringFromStringEncoding(reader)
	if err != nil {
		return "", nil, err
	}
	return key, &utils.Entry{Value: val, ExpiryTime: expiryTime}, nil
}

func getExpiryTime(reader *bufio.Reader) (time.Time, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return time.Time{}, err
	}
	switch b {
	case EXPIRY_SECONDS:
		data := make([]byte, EXPIRY_SECONDS_BYTE_LEN)
		n, err := reader.Read(data)
		if err != nil {
			return time.Time{}, err
		}
		if n != EXPIRY_SECONDS_BYTE_LEN {
			return time.Time{}, fmt.Errorf("expected to read %d bytes but got %d bytes instead", EXPIRY_SECONDS_BYTE_LEN, n)
		}
		return time.Unix(int64(binary.LittleEndian.Uint32(data)), 0), nil
	case EXPIRY_MILLISECONDS:
		data := make([]byte, EXPIRY_MILLISECONDS_BYTE_LEN)
		n, err := reader.Read(data)
		if err != nil {
			return time.Time{}, err
		}
		if n != EXPIRY_MILLISECONDS_BYTE_LEN {
			return time.Time{}, fmt.Errorf("expected to read %d bytes but got %d bytes instead", EXPIRY_MILLISECONDS_BYTE_LEN, n)
		}
		return time.UnixMilli(int64(binary.LittleEndian.Uint64(data))), nil

	default:
		return time.Time{}, fmt.Errorf("expected %q or %q, got %q", EXPIRY_SECONDS, EXPIRY_MILLISECONDS, b)
	}
}

func getStringFromStringEncoding(reader *bufio.Reader) (string, error) {
	len, err := getLengthFromStringEncoding(reader)
	if err != nil {
		return "", err
	}
	str := ""
	for range len {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		str += string(b)
	}
	return str, nil
}

func getChecksum(reader *bufio.Reader) (string, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return "", err
	}
	if b != CHECKSUM_OPCODE {
		return "", fmt.Errorf("expected %q, got %q", CHECKSUM_OPCODE, b)
	}
	checksum := ""
	for range CHECKSUM_LENGTH {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		checksum += string(b)
	}
	return checksum, nil
}

func getLengthFromStringEncoding(reader *bufio.Reader) (int, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return -1, err
	}
	firstTwoBits := (b & 0b11000000) >> 6
	switch firstTwoBits {
	case 0b00:
		return int(b & 0b00111111), nil
	case 0b01:
		nextb, err := reader.ReadByte()
		if err != nil {
			return -1, err
		}
		v := (b&0b00111111)<<8 | nextb
		return int(v), nil
	case 0b10:
		var bytesToRead int
		switch b {
		case 0x80:
			bytesToRead = 4
		case 0x81:
			bytesToRead = 8
		default:
			return -1, fmt.Errorf("invalid format of length encoding")
		}
		data := make([]byte, bytesToRead)
		n, err := reader.Read(data)
		if err != nil {
			return -1, err
		}
		if n != bytesToRead {
			return -1, fmt.Errorf("only read %d bytes", n)
		}
		if bytesToRead == 4 {
			return int(binary.BigEndian.Uint32(data)), nil
		}
		return int(binary.BigEndian.Uint64(data)), nil
	case 0b11:
		return -1, fmt.Errorf("special length encoding, ignore")
	}
	return -1, fmt.Errorf("error extracting first two bits. ")
}
