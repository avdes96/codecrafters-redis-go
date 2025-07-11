package command

import "github.com/codecrafters-io/redis-starter-go/app/utils"

type Command struct {
	CMD  string
	ARGS []string
}

type Context struct {
	CurrentDatabase int
	Store           map[int]map[string]utils.Entry
	ConfigParams    map[string]string
	ReplicationInfo *utils.ReplicationInfo
}

type CommandHandler interface {
	Handle(args []string, ctx *Context) []byte
}

func NewCommandRegistry() map[string]CommandHandler {
	m := make(map[string]CommandHandler)
	m["ping"] = &Ping{}
	m["echo"] = &Echo{}
	m["set"] = &Set{}
	m["get"] = &Get{}
	m["config"] = &Config{}
	m["keys"] = &Keys{}
	m["info"] = &Info{}
	return m
}
