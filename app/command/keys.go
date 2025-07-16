package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Keys struct{}

func (k *Keys) Handle(args []string, ctx *utils.Context) {
	if len(args) != 1 || args[0] != "*" {
		utils.WriteToConnection(ctx.Conn, []byte("Usage: KEYS *"))
		return
	}
	keys := []string{}
	if _, ok := ctx.Store[ctx.CurrentDatabase]; ok {
		for key := range ctx.Store[ctx.CurrentDatabase] {
			keys = append(keys, key)
		}
	}
	utils.WriteToConnection(ctx.Conn, protocol.ToArrayBulkStrings(keys))
}

func (k *Keys) IsWriteCommand() bool {
	return false
}
