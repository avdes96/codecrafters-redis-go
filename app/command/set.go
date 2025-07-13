package command

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Set struct{}

func (s *Set) Handle(args []string, ctx *Context) {
	if !(len(args) == 2 || len(args) == 4) {
		writeUsageString(ctx.Conn)
		return
	}
	var expTime time.Time
	if len(args) == 4 {
		if strings.ToLower(args[2]) != "px" {
			writeUsageString(ctx.Conn)
			return
		}
		expiryDelta, err := strconv.Atoi(args[3])
		if err != nil {
			writeUsageString(ctx.Conn)
			return
		}
		expTime = time.Now().Add(time.Millisecond * time.Duration(expiryDelta))
	}
	if _, ok := ctx.Store[ctx.CurrentDatabase]; !ok {
		ctx.Store[ctx.CurrentDatabase] = make(map[string]utils.Entry)
	}
	ctx.Store[ctx.CurrentDatabase][args[0]] = utils.Entry{Value: args[1], ExpiryTime: expTime}
	utils.WriteToConnection(ctx.Conn, protocol.OkResp())
}

const usageStr string = "Usage: SET <key> <value> [PX | milliseconds]"

func writeUsageString(conn net.Conn) {
	utils.WriteToConnection(conn, []byte(usageStr))
}

func (s *Set) IsWriteCommand() bool {
	return true
}
