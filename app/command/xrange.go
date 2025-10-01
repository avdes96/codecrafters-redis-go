package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/entry"
	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Xrange struct{}

const xrangeUsageStr string = "XADD key start end"
const streamNotExistError string = "ERR Stream does not exist"
const entryNotStreamError string = "ERR Entry not a stream"

func (x *Xrange) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	if len(args) != 3 {
		writeChan <- []byte(xrangeUsageStr)
		return
	}
	key, start, end := args[0], args[1], args[2]
	if _, ok := ctx.Store[ctx.CurrentDatabase]; !ok {
		writeChan <- protocol.ToError(streamNotExistError)
		return
	}
	e, ok := ctx.Store[ctx.CurrentDatabase][key]
	if !ok {
		writeChan <- protocol.ToError(streamNotExistError)
		return
	}
	s, ok := e.(*entry.Stream)
	if !ok {
		writeChan <- protocol.ToError(entryNotStreamError)
		return
	}
	items, err := s.GetDataFromRange(start, end)
	if err != nil {
		writeChan <- protocol.ToError(err.Error())
		return
	}
	writeChan <- items.Encoded()
}

func (x *Xrange) CanPropogateCommand(args []string) bool {
	return false
}
