package command

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Get struct{}

func (g *Get) Handle(args []string, ctx *Context) {
	if len(args) != 1 {
		utils.WriteToConnection(ctx.Conn, []byte("Usage: GET <key>"))
		return
	}
	entry, ok := ctx.Store[ctx.CurrentDatabase][args[0]]
	if !ok {
		utils.WriteToConnection(ctx.Conn, protocol.NullBulkString())
		return
	}
	if !entry.ExpiryTime.IsZero() && entry.ExpiryTime.Before(time.Now()) {
		delete(ctx.Store[ctx.CurrentDatabase], args[0])
		utils.WriteToConnection(ctx.Conn, protocol.NullBulkString())
		return
	}
	utils.WriteToConnection(ctx.Conn, protocol.ToBulkString(entry.Value))
}

func (g *Get) IsWriteCommand() bool {
	return false
}
