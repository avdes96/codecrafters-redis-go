package event

import (
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Event struct {
	Conn net.Conn
	Cmd  utils.Command
	Ctx  Context
}

type EventQueue struct {
	Queue  chan *Event
	Locked bool
	Mu     sync.RWMutex
}

func NewEventQueue() *EventQueue {
	return &EventQueue{
		Queue:  make(chan *Event, 100),
		Locked: false,
	}
}

func (eq *EventQueue) IsLocked() bool {
	eq.Mu.RLock()
	defer eq.Mu.RUnlock()
	return eq.Locked
}

func (eq *EventQueue) Lock() {
	eq.Mu.Lock()
	defer eq.Mu.Unlock()
	eq.Locked = true
}

func (eq *EventQueue) Unlock() {
	eq.Mu.Lock()
	defer eq.Mu.Unlock()
	eq.Locked = false
}

func (eq *EventQueue) Add(e *Event) {
	eq.Queue <- e
}

type Context struct {
	Conn            net.Conn
	ConnType        replication.ConnType
	CurrentDatabase int
	Store           map[int]map[string]utils.Entry
	ConfigParams    map[string]string
	ReplicationInfo *replication.ReplicationInfo
	EventQueue      *EventQueue
}
