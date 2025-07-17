package command

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Get struct{}

func (g *Get) Handle(args []string, ctx *utils.Context, writeChan chan []byte) {
	if len(args) != 1 {
		writeChan <- []byte("Usage: GET <key>")
		return
	}
	entry, ok := ctx.Store[ctx.CurrentDatabase][args[0]]
	if !ok {
		writeChan <- protocol.NullBulkString()
		return
	}
	if !entry.ExpiryTime.IsZero() && entry.ExpiryTime.Before(time.Now()) {
		delete(ctx.Store[ctx.CurrentDatabase], args[0])
		writeChan <- protocol.NullBulkString()
		return
	}
	writeChan <- protocol.ToBulkString(entry.Value)
}

func (g *Get) IsWriteCommand() bool {
	return false
}
