package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Echo struct{}

func (e *Echo) Handle(args []string) []byte {
	var ret string
	if len(args) == 1 {
		ret = args[0]
	} else {
		ret = "Usage: ECHO <message>"
	}
	return utils.ToBulkString(ret)
}
