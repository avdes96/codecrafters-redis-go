package entry

import (
	"fmt"
	"regexp"
	"strconv"
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

func (s *Stream) Type() string {
	return "stream"
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

func (s *Stream) Add(idStr string, field string, value string) (*streamID, string) {
	id, errMsg := s.validateID(idStr)
	if errMsg != "" {
		return nil, errMsg
	}
	s.lock.Lock()
	new := newStreamItem(id)
	s.topID = id
	s.data.ReplaceOrInsert(new)
	s.lock.Unlock()

	s.data.Get(new).(*StreamItem).AddField(field, value)
	return new.id, ""
}

func NewStreamID(m int, sn int) *streamID {
	return &streamID{
		millisecondsTime: m,
		sequenceNumber:   sn,
	}
}

var fullIDRe = regexp.MustCompile(`^(\d+)-(\d+)$`)
var partialIDRe = regexp.MustCompile(`^(\d+)-\*$`)

const invalidIDFormatStr string = "ERR id not in format <millisecondsTime>-<sequenceNumber>"
const invalidIDNotGreaterThanTopItem string = "ERR The ID specified in XADD is equal or smaller than the target stream top item"
const invalidIDNotGreaterThanZero string = "ERR The ID specified in XADD must be greater than 0-0"

func (s *Stream) validateID(id string) (*streamID, string) {
	if match := partialIDRe.FindStringSubmatch(id); match != nil {
		return s.validatePartialID(match[1])
	}
	if match := fullIDRe.FindStringSubmatch(id); match != nil {
		return s.validateFullID(match[1], match[2])
	}
	return nil, invalidIDFormatStr
}

func (s *Stream) validatePartialID(millisecondsTimeStr string) (*streamID, string) {
	millisecondsTime, err := strconv.Atoi(millisecondsTimeStr)
	if err != nil {
		return nil, invalidIDFormatStr
	}
	if millisecondsTime < 0 {
		return nil, invalidIDNotGreaterThanZero
	}
	if millisecondsTime < s.topID.millisecondsTime {
		return nil, invalidIDNotGreaterThanTopItem
	}
	if millisecondsTime == s.topID.millisecondsTime {
		return NewStreamID(millisecondsTime, s.topID.sequenceNumber+1), ""
	}
	sequenceNumber := 0
	if millisecondsTime == 0 {
		sequenceNumber = 1
	}
	return NewStreamID(millisecondsTime, sequenceNumber), ""
}

func (s *Stream) validateFullID(millisecondsTimeStr string, sequenceNumberStr string) (*streamID, string) {
	millisecondsTime, err := strconv.Atoi(millisecondsTimeStr)
	if err != nil {
		return nil, invalidIDFormatStr
	}
	sequenceNumber, err := strconv.Atoi(sequenceNumberStr)
	if err != nil {
		return nil, invalidIDFormatStr
	}

	if millisecondsTime < 0 || sequenceNumber < 0 || (millisecondsTime == 0 && sequenceNumber == 0) {
		return nil, invalidIDNotGreaterThanZero
	}

	if (millisecondsTime < s.topID.millisecondsTime) ||
		(millisecondsTime == s.topID.millisecondsTime && sequenceNumber <= s.topID.sequenceNumber) {
		return nil, invalidIDNotGreaterThanTopItem
	}

	return NewStreamID(millisecondsTime, sequenceNumber), ""
}
