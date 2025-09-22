package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/event"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/rdb"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type redisServer struct {
	listener        net.Listener
	clients         map[net.Conn]bool
	clientMutex     sync.RWMutex
	EventQueue      event.EventQueue
	syncList        *syncList
	parser          *protocol.Parser
	commandRegistry command.CommandRegistry
	store           map[int]map[string]utils.Entry
	configParams    map[string]string
	currentDatabase int
	replicationInfo *replication.ReplicationInfo
}

func New(configParams map[string]string, replInfo *replication.ReplicationInfo) (*redisServer, error) {
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
		EventQueue:      *event.NewEventQueue(),
		syncList:        &syncList{},
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
		go r.handleConnection(conn, replication.CONN_TYPE_CLIENT)
	}
}

func (r *redisServer) handleEvent(event *event.Event) {
	err := r.commandRegistry.Handle(event.Cmd, &event.Ctx)
	if err != nil {
		log.Printf("Error handling command: %s", err)
	}
}

func (r *redisServer) eventLoop() {
	for !r.syncList.replicaSyncDone {
	}
	for r.EventQueue.IsLocked() {
	}
	for event := range r.EventQueue.Queue {
		r.handleEvent(event)
	}
}

func (r *redisServer) handleConnection(conn net.Conn, t replication.ConnType) {
	r.clientMutex.Lock()
	r.clients[conn] = true
	r.clientMutex.Unlock()

	defer func() {
		conn.Close()
		r.clientMutex.Lock()
		delete(r.clients, conn)
		r.clientMutex.Unlock()
	}()

	reader := bufio.NewReader(conn)
	for {
		commandChan := make(chan utils.Command)
		replicaRespChan := make(chan string)
		go r.parser.Parse(reader, commandChan, replicaRespChan)
		ctx := event.Context{
			Conn:            conn,
			ConnType:        t,
			CurrentDatabase: r.currentDatabase,
			Store:           r.store,
			ConfigParams:    r.configParams,
			ReplicationInfo: r.replicationInfo,
			EventQueue:      &r.EventQueue,
		}
		r.handleChannels(commandChan, replicaRespChan, conn, ctx)
	}

}

func (r *redisServer) handleChannels(commandChan chan utils.Command, replicaRespChan chan string, conn net.Conn, ctx event.Context) {
	for {
		select {
		case cmd, ok := <-commandChan:
			if !ok {
				commandChan = nil
				break
			}
			r.EventQueue.Add(&event.Event{
				Conn: conn,
				Cmd:  cmd,
				Ctx:  ctx,
			})

		case resp, ok := <-replicaRespChan:
			if !ok {
				replicaRespChan = nil
				break
			}

			parts := strings.Split(strings.ToLower(resp), " ")
			if len(parts) >= 3 && parts[0] == "replconf" && parts[1] == "ack" {
				offset, err := strconv.Atoi(parts[2])
				if err != nil {
					continue
				}
				r.replicationInfo.UpdateReplicaOffset(conn, offset)
				break
			}
			r.syncList.update(resp)
		}
		if commandChan == nil && replicaRespChan == nil {
			return
		}
	}
}
