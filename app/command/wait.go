package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Wait struct{}

func (w *Wait) Handle(args []string, ctx *utils.Context, writeChan chan []byte) {
	numReplicas := len(ctx.ReplicationInfo.Replicas)
	writeChan <- protocol.ToRespInt(numReplicas)
}

func (w *Wait) IsWriteCommand() bool {
	return false
}
