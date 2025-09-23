package command

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/entry"
	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Get struct{}

func (g *Get) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	if len(args) != 1 {
		writeChan <- []byte("Usage: GET <key>")
		return
	}
	entry, ok := ctx.Store[ctx.CurrentDatabase][args[0]].(*entry.RedisString)
	if !ok {
		writeChan <- protocol.NullBulkString()
		return
	}
	if !entry.ExpiryTime().IsZero() && entry.ExpiryTime().Before(time.Now()) {
		delete(ctx.Store[ctx.CurrentDatabase], args[0])
		writeChan <- protocol.NullBulkString()
		return
	}
	writeChan <- protocol.ToBulkString(entry.Value())
}

func (g *Get) CanPropogateCommand(args []string) bool {
	return false
}
