package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Set struct{}

const usageStr string = "Usage: SET <key> <value> [PX | milliseconds]"

func (s *Set) Handle(args []string, ctx *Context) []byte {
	if !(len(args) == 2 || len(args) == 4) {
		return []byte(usageStr)
	}
	var expTime time.Time
	if len(args) == 4 {
		if strings.ToLower(args[2]) != "px" {
			return []byte(usageStr)
		}
		expiryDelta, err := strconv.Atoi(args[3])
		if err != nil {
			return []byte(usageStr)
		}
		expTime = time.Now().Add(time.Millisecond * time.Duration(expiryDelta))
	}
	ctx.Store[args[0]] = utils.Entry{Value: args[1], ExpiryTime: expTime}
	return protocol.OkResp()
}
