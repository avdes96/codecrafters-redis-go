package command

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Command struct {
	CMD  string
	ARGS []string
}

type Context struct {
	Conn            net.Conn
	CurrentDatabase int
	Store           map[int]map[string]utils.Entry
	ConfigParams    map[string]string
	ReplicationInfo *utils.ReplicationInfo
}

type CommandHandler interface {
	Handle(args []string, ctx *Context)
}

type CommandRegistry struct {
	Commands map[string]CommandHandler
}

func NewCommandRegistry() CommandRegistry {
	m := make(map[string]CommandHandler)
	m["ping"] = &Ping{}
	m["echo"] = &Echo{}
	m["set"] = &Set{}
	m["get"] = &Get{}
	m["config"] = &Config{}
	m["keys"] = &Keys{}
	m["info"] = &Info{}
	m["replconf"] = &Replconf{}
	m["psync"] = &Psync{}
	return CommandRegistry{Commands: m}
}

func (cr *CommandRegistry) Handle(cmd Command, ctx *Context) error {
	handler, ok := cr.Commands[cmd.CMD]
	if !ok {
		return fmt.Errorf("%s not a valid command", cmd.CMD)
	}
	handler.Handle(cmd.ARGS, ctx)
	return nil
}
