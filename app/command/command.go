package command

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type CommandHandler interface {
	Handle(args []string, ctx *event.Context, writeChan chan []byte)
	CanPropogateCommand([]string) bool
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
	m["wait"] = &Wait{}
	return CommandRegistry{Commands: m}
}

func (cr *CommandRegistry) Handle(cmd utils.Command, ctx *event.Context) error {
	cmdLower := strings.ToLower(cmd.CMD)
	handler, ok := cr.Commands[cmdLower]
	if !ok {
		return fmt.Errorf("%s not a valid command", cmd.CMD)
	}
	writeChan := make(chan []byte, 300)
	go func() {
		handler.Handle(cmd.ARGS, ctx, writeChan)
		close(writeChan)
	}()
	if canRespond(ctx, cmd) {
		for b := range writeChan {
			utils.WriteToConnection(ctx.Conn, b)
		}
	}
	switch ctx.ReplicationInfo.Role {
	case replication.ROLE_MASTER:
		propagateCommand(handler, cmd, ctx)
	case replication.ROLE_REPLICA:
		ctx.ReplicationInfo.IncrementServerOffset(cmd.ByteLen)
	}
	return nil
}

func propagateCommand(handler CommandHandler, cmd utils.Command, ctx *event.Context) {
	if !handler.CanPropogateCommand(cmd.ARGS) {
		return
	}
	msg := protocol.CommandAndArgsToBulkString(cmd.CMD, cmd.ARGS)
	ctx.ReplicationInfo.PropogateToReplicas(msg)
}

func canRespond(ctx *event.Context, cmd utils.Command) bool {
	if ctx.ConnType == replication.CONN_TYPE_CLIENT {
		return true
	}
	if strings.ToLower(cmd.CMD) == "replconf" && isGetackStar(cmd.ARGS) {
		return true
	}
	return false
}
