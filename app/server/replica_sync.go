package server

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

func (r *redisServer) SyncWithMaster() {
	defer close(r.replicaSyncDone)
	if r.replicationInfo.Role == utils.MASTER {
		return
	}
	conn, err := net.Dial("tcp", r.replicationInfo.MasterAddress)
	if err != nil {
		log.Printf("Error dialing master server: %s", err)
		return
	}

	if err = r.initiateConnection(conn); err != nil {
		log.Printf("Error initiating connection with master server: %s", err)
		return
	}

	if err = r.configureReplica(conn); err != nil {
		log.Printf("Error configuring replica: %s", err)
		return
	}

	if err = r.initialiseReplicationStream(conn); err != nil {
		log.Printf("Error initialising replication stream: %s", err)
		return
	}
}

func (r *redisServer) initiateConnection(conn net.Conn) error {
	_, err := conn.Write([]byte(protocol.ToArrayBulkStrings([]string{"PING"})))
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if string(buf[:n]) != "+PONG\r\n" {
		return fmt.Errorf("expected %s, got %s", "+PONG\r\n", string(buf[:n]))
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

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if string(buf[:n]) != "+OK\r\n" {
		return fmt.Errorf("expected %s, got %s", "+OK\r\n", string(buf[:n]))
	}

	_, err = conn.Write([]byte(protocol.ToArrayBulkStrings([]string{
		"REPLCONF", "capa", "psync2",
	})))
	if err != nil {
		return err
	}

	buf = make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		return err
	}
	if string(buf[:n]) != "+OK\r\n" {
		return fmt.Errorf("expected %s, got %s", "+OK\r\n", string(buf[:n]))
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

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(string(buf[:n]), fullresync) {
		return fmt.Errorf("expected resp to start with %s, got %s", fullresync, buf[:n])
	}
	go r.handleConnection(conn)

	return nil
}
