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
	replicationInfo utils.ReplicationInfo
}

func New(configParams map[string]string) (*redisServer, error) {
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
		replicationInfo: utils.ReplicationInfo{Role: utils.MASTER},
	}, nil
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
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing user input: %s", err)
			continue
		}
		output := r.commandRegistry[cmd.CMD].Handle(
			cmd.ARGS,
			&command.Context{Store: r.store, ConfigParams: r.configParams},
		)
		if _, err := conn.Write(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to connection %s", err.Error())
			continue
		}
	}
}
