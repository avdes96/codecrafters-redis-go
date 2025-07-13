package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/rdb"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type redisServer struct {
	listener        *net.Listener
	parser          *protocol.Parser
	commandRegistry map[string]command.CommandHandler
	store           map[int]map[string]utils.Entry
	configParams    map[string]string
	currentDatabase int
	replicationInfo *utils.ReplicationInfo
}

func New(configParams map[string]string, replInfo *utils.ReplicationInfo) (*redisServer, error) {
	portNum, ok := configParams["port"]
	if !ok {
		log.Fatal("Error fetching port")
	}
	address := fmt.Sprintf("0.0.0.0:%s", portNum)
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	p := protocol.NewParser()
	reg := command.NewCommandRegistry()

	rdbFile, _ := rdb.NewRdbFromFile(configParams["dir"], configParams["dbfilename"])
	var s map[int]map[string]utils.Entry
	if rdbFile == nil {
		s = make(map[int]map[string]utils.Entry)
	} else {
		s = rdbFile.Database
	}
	return &redisServer{
		listener:        &l,
		parser:          p,
		commandRegistry: reg,
		store:           s,
		configParams:    configParams,
		currentDatabase: 0,
		replicationInfo: replInfo,
	}, nil
}

func (r *redisServer) SyncWithMaster() {
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
	if !strings.HasPrefix(string(buf[:n]), "FULLRESYNC") {
		return fmt.Errorf("expected resp to start with %s, got %s", "FULLRESYNC", buf[:n])
	}

	return nil
}

func (r *redisServer) Run() error {
	defer (*r.listener).Close()
	for {
		conn, err := (*r.listener).Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %s", err.Error())
		}
		go r.handleConnection(conn)
	}
}

func (r *redisServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("Error reading from connection %s", err.Error())
			continue
		}

		userInput := buffer[:n]
		parsedCmd, parsedArgs, err := r.parser.ParseInputToCommandAndArgs(userInput)
		cmd := command.Command{CMD: parsedCmd, ARGS: parsedArgs}
		ctx := command.Context{
			Conn:            conn,
			CurrentDatabase: r.currentDatabase,
			Store:           r.store,
			ConfigParams:    r.configParams,
			ReplicationInfo: r.replicationInfo,
		}
		if err != nil {
			log.Printf("error parsing user input: %s", err)
			continue
		}
		r.commandRegistry[cmd.CMD].Handle(
			cmd.ARGS,
			&ctx,
		)
	}
}
