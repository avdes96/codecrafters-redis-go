package command

import (
	"fmt"
	"log"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/rdb"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Psync struct{}

func (p *Psync) Handle(args []string, ctx *Context) {
	ret := protocol.ToSimpleString(fmt.Sprintf("FULLRESYNC %s %s",
		ctx.ReplicationInfo.ReplicationId, strconv.Itoa(ctx.ReplicationInfo.Offset)))
	utils.WriteToConnection(ctx.Conn, ret)
	emptyRdbFile, err := rdb.EmptyRdbFile() // Assume rdb file is empty for this challenge
	if err != nil {
		log.Printf("Error loading empty rdb file: %s", err)
		return
	}
	utils.WriteToConnection(ctx.Conn, emptyRdbFile)
	ctx.ReplicationInfo.AddReplica(ctx.Conn)
}

func (p *Psync) IsWriteCommand() bool {
	return false
}
