package command

import (
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Replconf struct{}

func (r *Replconf) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	if isGetackStar(args) {
		offsetStr := strconv.Itoa(ctx.ReplicationInfo.GetServerOffset())
		strs := []string{"REPLCONF", "ACK", offsetStr}
		writeChan <- protocol.ToArrayBulkStrings(strs)
		return
	}
	writeChan <- protocol.OkResp()
}

func isGetackStar(args []string) bool {
	return len(args) == 2 && strings.ToLower(args[0]) == "getack" && args[1] == "*"
}

func (r *Replconf) CanPropogateCommand(args []string) bool {
	return false
}
