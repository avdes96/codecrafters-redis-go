package rdb

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

func TestFileParser(t *testing.T) {
	reader := getDummyFileReader()
	got, err := newRdbFromReader(reader)
	if err != nil {
		t.Errorf("Error creating rdb file: %s", err)
	}

	expectedDatabase := make(map[int]map[string]utils.Entry)
	expectedDatabase[0] = make(map[string]utils.Entry)
	expectedDatabase[0]["foobar"] = utils.Entry{Value: "bazqux"}
	expectedDatabase[0]["foo"] = utils.Entry{Value: "bar", ExpiryTime: time.UnixMilli(1713824559637)}
	expectedDatabase[0]["abcde"] = utils.Entry{Value: "wxyz", ExpiryTime: time.Unix(1714089298, 0)}
	expected := Rdb{
		header:   RdbHeader{magic: "REDIS", version: "0011"},
		metadata: map[string]string{"redis-ver": "6.0.16"},
		database: expectedDatabase,
		checksum: "ChEcKsUm",
	}

	if expected.header.magic != got.header.magic {
		t.Errorf("Expected header magic string: %s; got %s", expected.header.magic, got.header.magic)
	}

	if expected.header.version != got.header.version {
		t.Errorf("Expected header version string: %s; got %s", expected.header.version, got.header.version)
	}

	if len(expected.metadata) != len(got.metadata) {
		t.Errorf("Expected metadata is of len %d; got len %d", len(expected.metadata), len(got.metadata))
	}

	for key, value := range expected.metadata {
		gotVal, ok := got.metadata[key]
		if !ok {
			t.Errorf("%s not in got.metadata", key)
		}
		if value != gotVal {
			t.Errorf("Expected %s for %s, got %s", key, value, gotVal)
		}
	}
	if same, err := databasesEqual(expected.database, got.database); !same {
		t.Error(err.Error())
	}

	if expected.checksum != got.checksum {
		t.Errorf("Expected checksum %s; got %s", expected.checksum, got.checksum)
	}
}

func getDummyFileReader() *bufio.Reader {
	bs := []byte{}
	// Header
	bs = append(bs, []byte("REDIS0011")...)

	// Metadata
	bs = append(bs, METADATA_OPCODE)
	bs = append(bs, 0x09)
	bs = append(bs, []byte("redis-ver")...)
	bs = append(bs, 0x06)
	bs = append(bs, []byte("6.0.16")...)

	// Database
	bs = append(bs, DATABASE_OPCODE)
	bs = append(bs, 0x00)
	bs = append(bs, []byte{0xFB, 0x03, 0x02}...)

	bs = append(bs, 0x00)
	bs = append(bs, 0x06)
	bs = append(bs, []byte("foobar")...)
	bs = append(bs, 0x06)
	bs = append(bs, []byte("bazqux")...)

	bs = append(bs, 0xFC)
	bs = append(bs, []byte{0x15, 0x72, 0xE7, 0x07, 0x8F, 0x01, 0x00, 0x00}...)
	bs = append(bs, 0x00)
	bs = append(bs, 0x03)
	bs = append(bs, []byte("foo")...)
	bs = append(bs, 0x03)
	bs = append(bs, []byte("bar")...)

	bs = append(bs, 0xFD)
	bs = append(bs, []byte{0x52, 0xED, 0x2A, 0x66}...)
	bs = append(bs, 0x00)
	bs = append(bs, 0x05)
	bs = append(bs, []byte("abcde")...)
	bs = append(bs, 0x04)
	bs = append(bs, []byte("wxyz")...)

	// Checksum
	bs = append(bs, 0xFF)
	bs = append(bs, []byte("ChEcKsUm")...)

	return bufio.NewReader(bytes.NewReader(bs))
}

func databasesEqual(
	db1 map[int]map[string]utils.Entry,
	db2 map[int]map[string]utils.Entry,
) (bool, error) {
	if len(db1) != len(db2) {
		return false,
			fmt.Errorf("Expected database is of len %d; got len %d", len(db1), len(db2))
	}
	for key, dbSect1 := range db1 {
		dbSect2, ok := db2[key]
		if !ok {
			return false, fmt.Errorf("Key %d in %v, but not in %v", key, db1, db2)
		}
		if same, err := databaseSectionsEqual(dbSect1, dbSect2); !same {
			return false, err
		}

	}
	return true, nil
}

func databaseSectionsEqual(dbSect1 map[string]utils.Entry, dbSect2 map[string]utils.Entry) (bool, error) {
	if len(dbSect1) != len(dbSect2) {
		return false,
			fmt.Errorf("Expected section is of len %d; got len %d", len(dbSect1), len(dbSect2))
	}
	for key, entry1 := range dbSect1 {
		entry2, ok := dbSect1[key]
		if !ok {
			return false, fmt.Errorf("Key %s in %s, but not in %s", key, dbSect1, dbSect2)
		}
		if !entrySame(entry1, entry2) {
			return false, fmt.Errorf("entry %s does not match entry %s", entry1, entry2)
		}
	}
	return true, nil
}

func entrySame(a utils.Entry, b utils.Entry) bool {
	return a.Value == b.Value && a.ExpiryTime == b.ExpiryTime
}
