package command

import (
	"fmt"
	"log"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/rdb"
)

type Psync struct{}

func (p *Psync) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	ret := protocol.ToSimpleString(fmt.Sprintf("FULLRESYNC %s %s",
		ctx.ReplicationInfo.ReplicationId, strconv.Itoa(ctx.ReplicationInfo.GetServerOffset())))
	writeChan <- ret
	emptyRdbFile, err := rdb.EmptyRdbFile() // Assume rdb file is empty for this challenge
	if err != nil {
		log.Printf("Error loading empty rdb file: %s", err)
		return
	}
	writeChan <- emptyRdbFile
	ctx.ReplicationInfo.AddReplica(ctx.Conn)
}

func (p *Psync) CanPropogateCommand(args []string) bool {
	return false
}
