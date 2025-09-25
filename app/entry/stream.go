package entry

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/google/btree"
)

type streamID struct {
	millisecondsTime int
	sequenceNumber   int
}

func (id *streamID) String() string {
	return fmt.Sprintf("%d-%d", id.millisecondsTime, id.sequenceNumber)
}

type Stream struct {
	data  *btree.BTree
	topID *streamID
	lock  sync.RWMutex
}

type StreamItem struct {
	id     *streamID
	fields map[string]string
	lock   sync.RWMutex
}

func newStreamItem(id *streamID) *StreamItem {
	return &StreamItem{
		id:     id,
		fields: make(map[string]string),
	}
}

func (si *StreamItem) Less(than btree.Item) bool {
	thanAsStream := than.(*StreamItem)
	if si.id.millisecondsTime != thanAsStream.id.millisecondsTime {
		return si.id.millisecondsTime < thanAsStream.id.millisecondsTime
	}
	return si.id.sequenceNumber < thanAsStream.id.sequenceNumber
}

func (si *StreamItem) AddField(field string, value string) {
	si.lock.Lock()
	defer si.lock.Unlock()
	si.fields[field] = value
}

func NewStream() *Stream {
	return &Stream{
		data: btree.New(32),
		topID: &streamID{
			millisecondsTime: 0,
			sequenceNumber:   0,
		},
	}
}

func (s *Stream) Add(idStr string, field string, value string) string {
	id, errMsg := s.validateId(idStr)
	if errMsg != "" {
		return errMsg
	}
	s.lock.Lock()
	new := newStreamItem(id)
	s.topID = id
	s.data.ReplaceOrInsert(new)
	s.lock.Unlock()

	s.data.Get(new).(*StreamItem).AddField(field, value)
	return ""
}

func (s *Stream) Type() string {
	return "stream"
}

func NewStreamID(m int, sn int) *streamID {
	return &streamID{
		millisecondsTime: m,
		sequenceNumber:   sn,
	}
}

func (s *Stream) validateId(id string) (*streamID, string) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return nil, "ERR id not in format <millisecondsTime>-<sequenceNumber>"
	}
	millisecondsTime, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, "ERR id not in format <millisecondsTime>-<sequenceNumber>"
	}
	sequenceNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, "ERR id not in format <millisecondsTime>-<sequenceNumber>"
	}

	if millisecondsTime < 0 || sequenceNumber < 0 || (millisecondsTime == 0 && sequenceNumber == 0) {
		return nil, "ERR The ID specified in XADD must be greater than 0-0"
	}

	if (millisecondsTime < s.topID.millisecondsTime) ||
		(millisecondsTime == s.topID.millisecondsTime && sequenceNumber <= s.topID.sequenceNumber) {
		return nil, "ERR The ID specified in XADD is equal or smaller than the target stream top item"
	}

	return NewStreamID(millisecondsTime, sequenceNumber), ""
}
