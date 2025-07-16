package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/rdb"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type redisServer struct {
	listener        net.Listener
	clients         map[net.Conn]bool
	clientMutex     sync.RWMutex
	EventQueue      chan utils.Event
	replicaSyncDone chan struct{}
	parser          *protocol.Parser
	commandRegistry command.CommandRegistry
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
		listener:        l,
		clients:         make(map[net.Conn]bool),
		EventQueue:      make(chan utils.Event, 100),
		replicaSyncDone: make(chan struct{}),
		parser:          p,
		commandRegistry: reg,
		store:           s,
		configParams:    configParams,
		currentDatabase: 0,
		replicationInfo: replInfo,
	}, nil
}

func (r *redisServer) Run() error {
	defer r.listener.Close()

	go r.eventLoop()
	go r.SyncWithMaster()

	for {
		conn, err := r.listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %s", err.Error())
		}
		go r.handleConnection(conn)
	}
}

func (r *redisServer) handleEvent(event utils.Event) {
	err := r.commandRegistry.Handle(event.Cmd, &event.Ctx)
	if err != nil {
		log.Printf("Error handling command: %s", err)
	}
}

func (r *redisServer) eventLoop() {
	<-r.replicaSyncDone
	for event := range r.EventQueue {
		r.handleEvent(event)
	}
}

func (r *redisServer) handleConnection(conn net.Conn) {
	r.clientMutex.Lock()
	r.clients[conn] = true
	r.clientMutex.Unlock()

	defer func() {
		conn.Close()
		r.clientMutex.Lock()
		delete(r.clients, conn)
		r.clientMutex.Unlock()
	}()

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
		commandChan := make(chan utils.Command)
		go r.parser.Parse(userInput, commandChan)
		ctx := utils.Context{
			Conn:            conn,
			CurrentDatabase: r.currentDatabase,
			Store:           r.store,
			ConfigParams:    r.configParams,
			ReplicationInfo: r.replicationInfo,
		}
		for cmd := range commandChan {
			r.EventQueue <- utils.Event{
				Conn: conn,
				Cmd:  cmd,
				Ctx:  ctx,
			}
		}
	}
}
