package command

type Command struct {
	CMD  string
	ARGS []string
}

type Context struct {
	Store map[string]string
}

type CommandHandler interface {
	Handle(args []string, ctx *Context) []byte
}

func NewCommandRegistry() map[string]CommandHandler {
	m := make(map[string]CommandHandler)
	m["ping"] = &Ping{}
	m["echo"] = &Echo{}
	m["set"] = &Set{}
	m["get"] = &Get{}

	return m
}
