package entry

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
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
	data                          *btree.BTree
	dataLock                      sync.RWMutex
	topID                         *streamID
	highSequenceNumberPerTime     map[int]int
	highSequenceNumberPerTimeLock sync.RWMutex
}

func (s *Stream) Type() string {
	return "stream"
}

type KeyValue struct {
	Key   string
	Value string
}

type StreamItem struct {
	id     *streamID
	fields []*KeyValue // We need fields to be deterministic for testing
	lock   sync.RWMutex
}

func newStreamItem(id *streamID) *StreamItem {
	return &StreamItem{
		id:     id,
		fields: []*KeyValue{},
	}
}

func (si *StreamItem) Less(than btree.Item) bool {
	thanAsStream := than.(*StreamItem)
	if si.id.millisecondsTime != thanAsStream.id.millisecondsTime {
		return si.id.millisecondsTime < thanAsStream.id.millisecondsTime
	}
	return si.id.sequenceNumber < thanAsStream.id.sequenceNumber
}

func (si *StreamItem) AddField(field string, value string) *StreamItem {
	si.lock.Lock()
	defer si.lock.Unlock()
	si.fields = append(si.fields, &KeyValue{Key: field, Value: value})
	return si
}

func NewStream() *Stream {
	return &Stream{
		data: btree.New(32),
		topID: &streamID{
			millisecondsTime: 0,
			sequenceNumber:   0,
		},
		highSequenceNumberPerTime: make(map[int]int),
	}
}

func (s *Stream) Add(idStr string, field string, value string) (*streamID, error) {
	id, err := s.validateID(idStr)
	if err != nil {
		return nil, err
	}
	s.dataLock.Lock()
	new := newStreamItem(id).AddField(field, value)
	s.topID = id
	s.data.ReplaceOrInsert(new)
	s.highSequenceNumberPerTimeLock.Lock()
	s.highSequenceNumberPerTime[id.millisecondsTime] = id.sequenceNumber
	s.highSequenceNumberPerTimeLock.Unlock()
	s.dataLock.Unlock()

	return new.id, nil
}

func NewStreamID(m int, sn int) *streamID {
	return &streamID{
		millisecondsTime: m,
		sequenceNumber:   sn,
	}
}

type StreamRangeData []*StreamItem

func (s StreamRangeData) Encoded() []byte {
	ret := []byte{}
	ret = append(ret, []byte(fmt.Sprintf("*%d\r\n", len(s)))...)
	for _, item := range s {
		ret = append(ret, []byte("*2\r\n")...)
		ret = append(ret, protocol.ToBulkString(item.id.String())...)
		ret = append(ret, []byte(fmt.Sprintf("*%d\r\n", 2*len(item.fields)))...)
		for _, kv := range item.fields {
			ret = append(ret, protocol.ToBulkString(kv.Key)...)
			ret = append(ret, protocol.ToBulkString(kv.Value)...)
		}
	}
	return ret
}

var fullIDRe = regexp.MustCompile(`^(\d+)-(\d+)$`)
var partialIDRe = regexp.MustCompile(`^(\d+)-\*$`)

const invalidIDFormatStr string = "ERR id not in format <millisecondsTime>-<sequenceNumber|*> | *"
const invalidIDNotGreaterThanTopItem string = "ERR The ID specified in XADD is equal or smaller than the target stream top item"
const invalidIDNotGreaterThanZero string = "ERR The ID specified in XADD must be greater than 0-0"

func (s *Stream) validateID(id string) (*streamID, error) {
	if id == "*" {
		return s.generateID(), nil
	}
	if match := partialIDRe.FindStringSubmatch(id); match != nil {
		return s.validatePartialID(match[1])
	}
	if match := fullIDRe.FindStringSubmatch(id); match != nil {
		return s.validateFullID(match[1], match[2])
	}
	return nil, errors.New(invalidIDFormatStr)
}

func (s *Stream) generateID() *streamID {
	millisecondsTime := int(time.Now().UnixMilli())
	if millisecondsTime == s.topID.millisecondsTime {
		return NewStreamID(millisecondsTime, s.topID.sequenceNumber+1)
	}
	sequenceNumber := 0
	if millisecondsTime == 0 { // In edge case it is zero
		sequenceNumber = 1
	}
	return NewStreamID(millisecondsTime, sequenceNumber)
}

func (s *Stream) validatePartialID(millisecondsTimeStr string) (*streamID, error) {
	millisecondsTime, err := strconv.Atoi(millisecondsTimeStr)
	if err != nil {
		return nil, errors.New(invalidIDFormatStr)
	}
	if millisecondsTime < 0 {
		return nil, errors.New(invalidIDNotGreaterThanZero)
	}
	if millisecondsTime < s.topID.millisecondsTime {
		return nil, errors.New(invalidIDNotGreaterThanTopItem)
	}
	if millisecondsTime == s.topID.millisecondsTime {
		return NewStreamID(millisecondsTime, s.topID.sequenceNumber+1), nil
	}
	sequenceNumber := 0
	if millisecondsTime == 0 {
		sequenceNumber = 1
	}
	return NewStreamID(millisecondsTime, sequenceNumber), nil
}

func (s *Stream) validateFullID(millisecondsTimeStr string, sequenceNumberStr string) (*streamID, error) {
	millisecondsTime, sequenceNumber, err := validateIDFormat(millisecondsTimeStr, sequenceNumberStr)
	if err != nil {
		return nil, err
	}

	if (millisecondsTime < s.topID.millisecondsTime) ||
		(millisecondsTime == s.topID.millisecondsTime && sequenceNumber <= s.topID.sequenceNumber) {
		return nil, errors.New(invalidIDNotGreaterThanTopItem)
	}

	return NewStreamID(millisecondsTime, sequenceNumber), nil
}

func validateIDFormat(millisecondsTimeStr string, sequenceNumberStr string) (int, int, error) {
	millisecondsTime, err := strconv.Atoi(millisecondsTimeStr)
	if err != nil {
		return -1, -1, errors.New(invalidIDFormatStr)
	}
	sequenceNumber, err := strconv.Atoi(sequenceNumberStr)
	if err != nil {
		return -1, -1, errors.New(invalidIDFormatStr)
	}

	if millisecondsTime < 0 || sequenceNumber < 0 || (millisecondsTime == 0 && sequenceNumber == 0) {
		return -1, -1, errors.New(invalidIDNotGreaterThanZero)
	}
	return millisecondsTime, sequenceNumber, nil
}

func (s *Stream) GetDataFromRange(startStr string, endStr string) (StreamRangeData, error) {
	startID, err := s.createStartStreamID(startStr)
	if err != nil {
		return nil, err
	}
	endID, err := s.createEndStreamID(endStr)
	if err != nil {
		return nil, err
	}
	var result []*StreamItem
	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	s.data.AscendGreaterOrEqual(newStreamItem(startID), func(item btree.Item) bool {
		if newStreamItem(endID).Less(item) {
			return false
		}
		result = append(result, item.(*StreamItem))
		return true
	})
	return result, nil
}

func (s *Stream) createStartStreamID(startStr string) (*streamID, error) {
	match := fullIDRe.FindStringSubmatch(startStr)
	if match != nil {
		millisecondsTime, sequenceNum, err := validateIDFormat(match[1], match[2])
		if err != nil {
			return nil, err
		}
		return NewStreamID(millisecondsTime, sequenceNum), nil
	}
	if ms, err := strconv.Atoi(startStr); err == nil {
		return NewStreamID(ms, 0), nil
	}
	return nil, errors.New(invalidIDFormatStr)
}

func (s *Stream) createEndStreamID(endStr string) (*streamID, error) {
	match := fullIDRe.FindStringSubmatch(endStr)
	if match != nil {
		millisecondsTime, sequenceNum, err := validateIDFormat(match[1], match[2])
		if err != nil {
			return nil, err
		}
		return NewStreamID(millisecondsTime, sequenceNum), nil
	}
	if ms, err := strconv.Atoi(endStr); err == nil {
		if sn, ok := s.highSequenceNumberPerTime[ms]; ok {
			return NewStreamID(ms, sn), nil
		}
		return NewStreamID(ms, 0), nil
	}
	return nil, errors.New(invalidIDFormatStr)
}
