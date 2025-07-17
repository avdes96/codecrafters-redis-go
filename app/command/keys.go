package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Keys struct{}

func (k *Keys) Handle(args []string, ctx *utils.Context, writeChan chan []byte) {
	if len(args) != 1 || args[0] != "*" {
		writeChan <- []byte("Usage: KEYS *")
		return
	}
	keys := []string{}
	if _, ok := ctx.Store[ctx.CurrentDatabase]; ok {
		for key := range ctx.Store[ctx.CurrentDatabase] {
			keys = append(keys, key)
		}
	}
	writeChan <- protocol.ToArrayBulkStrings(keys)
}

func (k *Keys) IsWriteCommand() bool {
	return false
}
