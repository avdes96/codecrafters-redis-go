package command

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Get struct{}

func (g *Get) Handle(args []string, ctx *Context) []byte {
	if len(args) != 1 {
		return []byte("Usage: GET <key>")
	}
	entry, ok := ctx.Store[args[0]]
	if !ok {
		return utils.NullBulkString()
	}
	if !entry.ExpiryTime.IsZero() && entry.ExpiryTime.Before(time.Now()) {
		delete(ctx.Store, args[0])
		return utils.NullBulkString()
	}
	return utils.ToBulkString(entry.Value)
}
