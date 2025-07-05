package server

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

type redisServer struct {
	listener        *net.Listener
	parser          *parser.Parser
	commandRegistry map[string]command.CommandHandler
	store           map[string]string
}

func New() (*redisServer, error) {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		return nil, err
	}
	p := parser.NewParser()
	reg := command.NewCommandRegistry()
	s := make(map[string]string)
	return &redisServer{
		listener:        &l,
		parser:          p,
		commandRegistry: reg,
		store:           s,
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
		cmd, err := r.parser.ParseInputToCommand(userInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing user input: %s", err)
			continue
		}
		output := r.commandRegistry[cmd.CMD].Handle(cmd.ARGS, &command.Context{Store: r.store})
		if _, err := conn.Write(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to connection %s", err.Error())
			continue
		}
	}
}
