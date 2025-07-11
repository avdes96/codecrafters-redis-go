package command

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Get struct{}

func (g *Get) Handle(args []string, ctx *Context) []byte {
	if len(args) != 1 {
		return []byte("Usage: GET <key>")
	}
	entry, ok := ctx.Store[ctx.CurrentDatabase][args[0]]
	if !ok {
		return protocol.NullBulkString()
	}
	if !entry.ExpiryTime.IsZero() && entry.ExpiryTime.Before(time.Now()) {
		delete(ctx.Store[ctx.CurrentDatabase], args[0])
		return protocol.NullBulkString()
	}
	return protocol.ToBulkString(entry.Value)
}
