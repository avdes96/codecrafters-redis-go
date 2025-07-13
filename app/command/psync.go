package command

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Psync struct{}

func (p *Psync) Handle(args []string, ctx *Context) {
	ret := protocol.ToSimpleString(fmt.Sprintf("FULLRESYNC %s %s",
		ctx.ReplicationInfo.ReplicationId, strconv.Itoa(ctx.ReplicationInfo.Offset)))
	utils.WriteToConnection(ctx.Conn, ret)
}
