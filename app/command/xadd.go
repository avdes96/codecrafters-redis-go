package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/entry"
	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Xadd struct{}

const xaddUsageStr string = "XADD key <* | id> field value [field value ...]"

func (x *Xadd) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	if len(args) < 4 {
		writeChan <- []byte(xaddUsageStr)
		return
	}
	key, id, field, value := args[0], args[1], args[2], args[3]
	if _, ok := ctx.Store[ctx.CurrentDatabase]; !ok {
		ctx.Store[ctx.CurrentDatabase] = make(map[string]entry.Entry)
	}
	e, ok := ctx.Store[ctx.CurrentDatabase][key]
	if !ok {
		e = entry.NewStream()
		ctx.Store[ctx.CurrentDatabase][key] = e
	}
	stream, ok := e.(*entry.Stream)
	if !ok {
		writeChan <- protocol.ToError("WRONGTYPE entry at key is not a stream")
		return
	}
	stream.Add(id, field, value)
	writeChan <- protocol.ToBulkString(id)
}

func (x *Xadd) CanPropogateCommand(args []string) bool {
	return true
}
