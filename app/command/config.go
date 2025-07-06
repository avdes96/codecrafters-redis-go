package command

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Config struct{}

func (c *Config) Handle(args []string, ctx *Context) []byte {
	switch strings.ToLower(args[0]) {
	case "get":
		if len(args) < 2 {
			return []byte("Usage: CONFIG GET <config_param> <config_param>...")
		}
		ret := []string{}
		for _, key := range args[1:] {
			if val, ok := ctx.ConfigParams[key]; ok {
				ret = append(ret, key)
				ret = append(ret, val)
			}
		}
		return utils.ToArrayBulkStrings(ret)
	default:
		return []byte("Available CONFIG commands: GET")
	}
}
