package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Replconf struct{}

func (r *Replconf) Handle(args []string, ctx *utils.Context) {
	utils.WriteToConnection(ctx.Conn, protocol.OkResp())
}

func (r *Replconf) IsWriteCommand() bool {
	return false
}
