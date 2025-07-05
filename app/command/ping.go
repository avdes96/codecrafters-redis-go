package command

type Ping struct{}

func (p *Ping) Handle(args []string, ctx *Context) []byte {
	return []byte("+PONG\r\n")
}
