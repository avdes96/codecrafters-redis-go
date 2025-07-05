package command

type Ping struct{}

func (p *Ping) Handle(args []string) []byte {
	return []byte("+PONG\r\n")
}
