package command

import "github.com/codecrafters-io/redis-starter-go/app/utils"

type Ping struct{}

func (p *Ping) Handle(args []string, ctx *Context) {
	utils.WriteToConnection(ctx.Conn, []byte("+PONG\r\n"))
}

func (p *Ping) IsWriteCommand() bool {
	return false
}
