package command

import "github.com/codecrafters-io/redis-starter-go/app/utils"

type Get struct{}

func (g *Get) Handle(args []string, ctx *Context) []byte {
	if len(args) != 1 {
		return []byte("Usage: GET <key>")
	}
	if val, ok := ctx.Store[args[0]]; ok {
		return utils.ToBulkString(val)
	}
	return utils.OkResp()
}
