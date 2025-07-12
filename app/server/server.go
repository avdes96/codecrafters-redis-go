package server

import (
	"fmt"
	"io"
	"net"
	"os"

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
		os.Exit(1)
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
		return
	}

	err = r.initiateConnection(conn)
	if err != nil {
		return
	}

	err = r.configureReplica(conn)
	if err != nil {
		return
	}
}

func (r *redisServer) initiateConnection(conn net.Conn) error {
	_, err := conn.Write([]byte(protocol.ToArrayBulkStrings([]string{"PING"})))
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
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

	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		return err
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
	return nil
}

func (r *redisServer) Run() error {
	defer (*r.listener).Close()
	for {
		conn, err := (*r.listener).Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error reading from connection %s", err.Error())
			continue
		}

		userInput := buffer[:n]
		parsedCmd, parsedArgs, err := r.parser.ParseInputToCommandAndArgs(userInput)
		cmd := command.Command{CMD: parsedCmd, ARGS: parsedArgs}
		ctx := command.Context{
			CurrentDatabase: r.currentDatabase,
			Store:           r.store,
			ConfigParams:    r.configParams,
			ReplicationInfo: r.replicationInfo,
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing user input: %s", err)
			continue
		}
		output := r.commandRegistry[cmd.CMD].Handle(
			cmd.ARGS,
			&ctx,
		)
		if _, err := conn.Write(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to connection %s", err.Error())
			continue
		}
	}
}
