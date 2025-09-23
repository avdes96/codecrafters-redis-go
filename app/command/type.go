package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

type Type struct{}

func (t *Type) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	if len(args) != 1 {
		writeChan <- []byte("Usage: TYPE <key>")
		return
	}
	if val, ok := ctx.Store[ctx.CurrentDatabase][args[0]]; ok {
		writeChan <- []byte(protocol.ToSimpleString(val.Type()))
		return
	}
	writeChan <- []byte(protocol.ToSimpleString("none"))
}

func (t *Type) CanPropogateCommand(args []string) bool {
	return false
}
