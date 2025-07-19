package utils

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

type Event struct {
	Conn net.Conn
	Cmd  Command
	Ctx  Context
}

type Context struct {
	Conn            net.Conn
	ConnType        ConnType
	CurrentDatabase int
	Store           map[int]map[string]Entry
	ConfigParams    map[string]string
	ReplicationInfo *ReplicationInfo
}

type Entry struct {
	Value      string
	ExpiryTime time.Time
}

type Command struct {
	CMD     string
	ARGS    []string
	ByteLen int
}

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
	Role          role
	ReplicationId string
	Offset        int
	MasterAddress string
	Replicas      map[net.Conn]bool
}

const replicationIdLen int = 40

func NewReplicationInfo(masterAddress string) *ReplicationInfo {
	role := ROLE_REPLICA
	formattedAddress := formatAddress(masterAddress)
	if formattedAddress == "" {
		role = ROLE_MASTER
	}
	return &ReplicationInfo{
		Role:          role,
		ReplicationId: randomAlphanumericString(replicationIdLen),
		Offset:        0,
		MasterAddress: formattedAddress,
		Replicas:      make(map[net.Conn]bool),
	}
}

func (r *ReplicationInfo) AddReplica(c net.Conn) {
	if r.Role != ROLE_MASTER {
		return
	}
	r.Replicas[c] = true
}

func formatAddress(a string) string {
	parts := strings.Split(a, " ")
	if len(parts) != 2 {
		return ""
	}
	return fmt.Sprintf("%s:%s", parts[0], parts[1])
}

const options string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomAlphanumericString(n int) string {
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range n {
		b[i] = options[gen.Intn(len(options))]
	}
	return string(b)
}

func SlicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func WriteToConnection(conn net.Conn, b []byte) {
	if _, err := conn.Write(b); err != nil {
		log.Printf("Error writing to connection %s", err.Error())
	}
}
