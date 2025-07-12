package command

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Psync struct{}

func (p *Psync) Handle(args []string, ctx *Context) []byte {
	return protocol.ToSimpleString(fmt.Sprintf("FULLRESYNC %s %s",
		ctx.ReplicationInfo.ReplicationId, strconv.Itoa(ctx.ReplicationInfo.Offset)))
}
