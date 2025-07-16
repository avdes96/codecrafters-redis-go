package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Echo struct{}

func (e *Echo) Handle(args []string, ctx *utils.Context) {
	var ret string
	if len(args) == 1 {
		ret = args[0]
	} else {
		ret = "Usage: ECHO <message>"
	}
	utils.WriteToConnection(ctx.Conn, protocol.ToBulkString(ret))
}

func (e *Echo) IsWriteCommand() bool {
	return false
}
