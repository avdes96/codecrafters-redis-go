package command

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Info struct{}

func (i *Info) Handle(args []string, ctx *Context) []byte {
	role := ctx.ReplicationInfo.Role.String()
	return protocol.ToBulkString(fmt.Sprintf("role:%s", role))
}
