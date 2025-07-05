package command

type Command struct {
	CMD  string
	ARGS []string
}

type CommandHandler interface {
	Handle(args []string) []byte
}

func NewCommandRegistry() map[string]CommandHandler {
	m := make(map[string]CommandHandler)
	m["ping"] = &Ping{}
	m["echo"] = &Echo{}

	return m
}
