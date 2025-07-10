package command

import "github.com/codecrafters-io/redis-starter-go/app/protocol"

type Keys struct{}

func (k *Keys) Handle(args []string, ctx *Context) []byte {
	if len(args) != 1 || args[0] != "*" {
		return []byte("Usage: KEYS *")
	}
	keys := []string{}
	if _, ok := ctx.Store[ctx.currentDatabase]; ok {
		for key := range ctx.Store[ctx.currentDatabase] {
			keys = append(keys, key)
		}
	}
	return protocol.ToArrayBulkStrings(keys)
}
