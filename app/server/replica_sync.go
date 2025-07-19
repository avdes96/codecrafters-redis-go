package server

import (
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

func (r *redisServer) SyncWithMaster() {
	if r.replicationInfo.Role == utils.ROLE_MASTER {
		r.syncList.completeSync()
		return
	}
	conn, err := net.Dial("tcp", r.replicationInfo.MasterAddress)
	if err != nil {
		log.Printf("Error dialing master server: %s", err)
		return
	}

	go r.handleConnection(conn, utils.CONN_TYPE_REPLICA)
	if err = r.initiateConnection(conn); err != nil {
		log.Printf("Error initiating connection with master server: %s", err)
		return
	}
	for !r.syncList.pingDone {
	}

	if err = r.configureReplica(conn); err != nil {
		log.Printf("Error configuring replica: %s", err)
		return
	}
	for !r.syncList.firstOkDone {
	}
	for !r.syncList.secondOkDone {
	}

	if err = r.initialiseReplicationStream(conn); err != nil {
		log.Printf("Error initialising replication stream: %s", err)
		return
	}
	for !r.syncList.fullresyncDone {
	}
	r.syncList.completeSync()
}

func (r *redisServer) initiateConnection(conn net.Conn) error {
	_, err := conn.Write([]byte(protocol.ToArrayBulkStrings([]string{"PING"})))
	if err != nil {
		return err
	}
	return nil
}

func (r *redisServer) configureReplica(conn net.Conn) error {
	port := r.configParams["port"]
	if port == "" {
		return nil
	}
	_, err := conn.Write([]byte(protocol.ToArrayBulkStrings([]string{
		"REPLCONF", "listening-port", port,
	})))
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(protocol.ToArrayBulkStrings([]string{
		"REPLCONF", "capa", "psync2",
	})))
	if err != nil {
		return err
	}
	return nil
}

const fullresync string = "+FULLRESYNC"

func (r *redisServer) initialiseReplicationStream(conn net.Conn) error {
	_, err := conn.Write([]byte(protocol.ToArrayBulkStrings([]string{
		"PSYNC", "?", "-1",
	})))
	if err != nil {
		return err
	}

	return nil
}

type syncList struct {
	pingDone        bool
	firstOkDone     bool
	secondOkDone    bool
	fullresyncDone  bool
	replicaSyncDone bool
}

func (s *syncList) update(response string) {
	if s.replicaSyncDone {
		return
	}
	respLower := strings.ToLower(response)
	if respLower == "pong" {
		s.pingDone = true
	}
	if !s.pingDone {
		return
	}
	if respLower == "ok" {
		if !s.firstOkDone {
			s.firstOkDone = true
			return
		}
		s.secondOkDone = true
	}
	if !(s.firstOkDone && s.secondOkDone) {
		return
	}
	if strings.HasPrefix(respLower, "fullresync") {
		s.fullresyncDone = true
	}
	if !s.fullresyncDone {
		return
	}

	s.completeSync()
}

func (s *syncList) completeSync() {
	s.replicaSyncDone = true
}
