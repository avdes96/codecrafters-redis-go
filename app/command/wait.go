package command

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
)

type Wait struct{}

func (w *Wait) Handle(args []string, ctx *event.Context, writeChan chan []byte) {
	numReplicas := len(ctx.ReplicationInfo.Replicas)
	if numReplicas == 0 {
		writeChan <- protocol.ToRespInt(0)
		return
	}
	if len(args) != 2 {
		writeChan <- []byte("Usage: WAIT replicasToRespond timeout")
		return
	}
	replicasToRespond, err := strconv.Atoi(args[0])
	if err != nil {
		writeChan <- []byte("Error: WAIT; unable to convert replicasToRespond to an int")
		return
	}
	timeout, err := strconv.Atoi(args[1])
	if err != nil {
		writeChan <- []byte("Error: WAIT; unable to convert timeout to an int")
		return
	}
	ctx.EventQueue.Lock()
	defer ctx.EventQueue.Unlock()
	wait(timeout, replicasToRespond, ctx, writeChan)
}

func (w *Wait) CanPropogateCommand(args []string) bool {
	return false
}

func wait(timeoutInMilliseconds int, n int, ctx *event.Context, writeChan chan []byte) {
	expectedOffset := ctx.ReplicationInfo.GetServerOffset()
	if count := numberInSync(expectedOffset, ctx.ReplicationInfo); count >= min(n, len(ctx.ReplicationInfo.Replicas)) {
		writeChan <- protocol.ToRespInt(count)
		return
	}

	ctx.ReplicationInfo.PropogateToReplicas(protocol.ToArrayBulkStrings([]string{"REPLCONF", "GETACK", "*"}))

	count := 0
	pollChan := make(chan bool, 100)
	var wg sync.WaitGroup

	waitTime := time.Duration(timeoutInMilliseconds) * time.Millisecond
	timeoutCtx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()

	for replica := range ctx.ReplicationInfo.Replicas {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pollReplica(expectedOffset, ctx.ReplicationInfo, replica, timeoutCtx, pollChan)
		}()
	}
	completedPolls := 0
	for isInSync := range pollChan {
		if isInSync {
			count += 1
			if count == n {
				break
			}
		}
		completedPolls++
		if completedPolls == len(ctx.ReplicationInfo.Replicas) {
			break
		}
	}
	cancel()
	wg.Wait()
	close(pollChan)
	writeChan <- protocol.ToRespInt(numberInSync(expectedOffset, ctx.ReplicationInfo))
}

func pollReplica(expectedOffset int, ri *replication.ReplicationInfo, replica net.Conn, timeoutCtx context.Context, pollChan chan bool) {
	for {
		select {
		case <-timeoutCtx.Done():
			pollChan <- false
			return
		default:
			if ri.GetReplicaOffset(replica) == expectedOffset {
				pollChan <- true
				return
			}
		}
	}

}

func numberInSync(expectedOffset int, ri *replication.ReplicationInfo) int {
	count := 0
	for replica := range ri.Replicas {
		if ri.GetReplicaOffset(replica) == expectedOffset {
			count++
		}
	}
	return count
}
