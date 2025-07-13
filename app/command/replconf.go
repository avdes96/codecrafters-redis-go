package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Replconf struct{}

func (k *Replconf) Handle(args []string, ctx *Context) {
	utils.WriteToConnection(ctx.Conn, protocol.OkResp())
}
