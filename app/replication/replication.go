package replication

import (
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type role int

const (
	ROLE_MASTER role = iota
	ROLE_REPLICA
)

type ConnType int

const (
	CONN_TYPE_CLIENT ConnType = iota
	CONN_TYPE_REPLICA
)

func (r role) String() string {
	// "slave" included as this version of codecrafters does not use updated renaming of "slave" to "replica"
	return [...]string{"master", "slave"}[r]
}

type ReplicationInfo struct {
	Role               role
	ReplicationId      string
	ServerOffset       int
	ServerOffsetLock   sync.RWMutex
	MasterAddress      string
	Replicas           map[net.Conn]bool
	ReplicasLock       sync.RWMutex
	ReplicaOffsets     map[net.Conn]int
	ReplicaOffsetsLock sync.RWMutex
}

func (ri *ReplicationInfo) IncrementServerOffset(inc int) {
	ri.ServerOffsetLock.Lock()
	defer ri.ServerOffsetLock.Unlock()
	ri.ServerOffset += inc
}

func (ri *ReplicationInfo) GetServerOffset() int {
	ri.ServerOffsetLock.RLock()
	defer ri.ServerOffsetLock.RUnlock()
	return ri.ServerOffset
}

func (ri *ReplicationInfo) UpdateReplicaOffset(replica net.Conn, val int) {
	if val == ri.GetReplicaOffset(replica) {
		return
	}
	ri.ReplicaOffsetsLock.Lock()
	defer ri.ReplicaOffsetsLock.Unlock()
	ri.ReplicaOffsets[replica] = val
}

func (ri *ReplicationInfo) GetReplicaOffset(replica net.Conn) int {
	ri.ReplicaOffsetsLock.RLock()
	defer ri.ReplicaOffsetsLock.RUnlock()
	val, ok := ri.ReplicaOffsets[replica]
	if ok {
		return val
	}
	return -1
}

const replicationIdLen int = 40

func NewReplicationInfo(masterAddress string) *ReplicationInfo {
	role := ROLE_REPLICA
	formattedAddress := utils.FormatAddress(masterAddress)
	if formattedAddress == "" {
		role = ROLE_MASTER
	}
	return &ReplicationInfo{
		Role:           role,
		ReplicationId:  utils.RandomAlphanumericString(replicationIdLen),
		ServerOffset:   0,
		MasterAddress:  formattedAddress,
		Replicas:       make(map[net.Conn]bool),
		ReplicaOffsets: make(map[net.Conn]int),
	}
}

func (ri *ReplicationInfo) AddReplica(c net.Conn) {
	if ri.Role != ROLE_MASTER {
		return
	}
	ri.ReplicasLock.Lock()
	defer ri.ReplicasLock.Unlock()
	ri.ReplicaOffsetsLock.Lock()
	defer ri.ReplicaOffsetsLock.Unlock()
	ri.Replicas[c] = true
	ri.ReplicaOffsets[c] = 0
}

func (ri *ReplicationInfo) PropogateToReplicas(b []byte) {
	if ri.Role != ROLE_MASTER {
		return
	}
	ri.IncrementServerOffset(len(b))
	for replica := range ri.Replicas {
		utils.WriteToConnection(replica, b)
	}
}
