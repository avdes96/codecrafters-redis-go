package command

import "github.com/codecrafters-io/redis-starter-go/app/protocol"

type Replconf struct{}

func (k *Replconf) Handle(args []string, ctx *Context) []byte {
	return protocol.OkResp()
}
