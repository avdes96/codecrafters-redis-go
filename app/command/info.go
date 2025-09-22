package command

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Info struct{}

func (i *Info) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	role := ctx.ReplicationInfo.Role.String()
	offset := strconv.Itoa(ctx.ReplicationInfo.GetServerOffset())
	id := ctx.ReplicationInfo.ReplicationId
	output := fmt.Sprintf("role:%s\r\nmaster_repl_offset:%s\r\nmaster_replid:%s", role, offset, id)
	writeChan <- protocol.ToBulkString(output)
}

func (i *Info) CanPropogateCommand(args []string) bool {
	return false
}
