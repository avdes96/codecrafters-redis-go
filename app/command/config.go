package command

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Config struct{}

func (c *Config) Handle(args []string, ctx *utils.Context, writeChan chan []byte) {
	var ret []byte
	switch strings.ToLower(args[0]) {
	case "get":
		if len(args) < 2 {
			ret = []byte("Usage: CONFIG GET <config_param> <config_param>...")
			break
		}
		strs := []string{}
		for _, key := range args[1:] {
			if val, ok := ctx.ConfigParams[key]; ok {
				strs = append(strs, key)
				strs = append(strs, val)
			}
		}
		ret = protocol.ToArrayBulkStrings(strs)
	default:
		ret = []byte("Available CONFIG commands: GET")
	}
	writeChan <- ret
}

func (c *Config) IsWriteCommand() bool {
	return false
}
