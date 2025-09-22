package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Echo struct{}

func (e *Echo) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	var ret string
	if len(args) == 1 {
		ret = args[0]
	} else {
		ret = "Usage: ECHO <message>"
	}
	writeChan <- protocol.ToBulkString(ret)
}

func (e *Echo) CanPropogateCommand(args []string) bool {
	return false
}
