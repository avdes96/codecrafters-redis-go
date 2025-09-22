package command

import "github.com/codecrafters-io/redis-starter-go/app/event"

type Ping struct{}

func (p *Ping) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	writeChan <- []byte("+PONG\r\n")
}

func (p *Ping) CanPropogateCommand(args []string) bool {
	return false
}
