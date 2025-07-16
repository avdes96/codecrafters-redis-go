package command

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type CommandHandler interface {
	Handle(args []string, ctx *utils.Context)
	IsWriteCommand() bool
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

func (cr *CommandRegistry) Handle(cmd utils.Command, ctx *utils.Context) error {
	handler, ok := cr.Commands[cmd.CMD]
	if !ok {
		return fmt.Errorf("%s not a valid command", cmd.CMD)
	}
	handler.Handle(cmd.ARGS, ctx)
	if ctx.ReplicationInfo.Role != utils.MASTER {
		return nil
	}
	if handler.IsWriteCommand() {
		for replica := range ctx.ReplicationInfo.Replicas {
			replica.Write(protocol.CommandAndArgsToBulkString(cmd.CMD, cmd.ARGS))
		}
	}
	return nil
}
