package command

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type CommandHandler interface {
	Handle(args []string, ctx *utils.Context, writeChan chan []byte)
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
	writeChan := make(chan []byte, 100)
	go func() {
		handler.Handle(cmd.ARGS, ctx, writeChan)
		close(writeChan)
	}()
	if canWrite(ctx) {
		for b := range writeChan {
			utils.WriteToConnection(ctx.Conn, b)
		}
	}

	propagateCommands(handler, cmd, ctx)
	return nil
}

func propagateCommands(handler CommandHandler, cmd utils.Command, ctx *utils.Context) {
	if ctx.ReplicationInfo.Role != utils.MASTER {
		return
	}
	if handler.IsWriteCommand() {
		for replica := range ctx.ReplicationInfo.Replicas {
			replica.Write(protocol.CommandAndArgsToBulkString(cmd.CMD, cmd.ARGS))
		}
	}
}

func canWrite(ctx *utils.Context) bool {
	if ctx.ReplicationInfo.Role == utils.MASTER {
		return true
	}
	return false
}
