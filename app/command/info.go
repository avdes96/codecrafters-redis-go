package command

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Info struct{}

func (i *Info) Handle(args []string, ctx *Context) []byte {
	role := ctx.ReplicationInfo.Role.String()
	offset := strconv.Itoa(ctx.ReplicationInfo.Offset)
	id := ctx.ReplicationInfo.ReplicationId
	output := fmt.Sprintf("role:%s\r\nmaster_repl_offset:%s\r\nmaster_replid:%s", role, offset, id)
	return protocol.ToBulkString(output)
}
