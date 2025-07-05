package command

import "github.com/codecrafters-io/redis-starter-go/app/utils"

type Set struct{}

func (s *Set) Handle(args []string, ctx *Context) []byte {
	if len(args) < 2 {
		return []byte("Usage: SET <key> <value>")
	}
	ctx.Store[args[0]] = args[1]
	return utils.OkResp()
}
