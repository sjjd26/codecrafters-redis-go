package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type Echo struct {
	Message string
}

func NewEcho(args []string) (Command, error) {
	if len(args) == 0 {
		return nil, ErrNotEnoughArgs
	} else if len(args) > 1 {
		return nil, ErrTooManyArgs
	}

	return Echo{Message: args[0]}, nil
}

func (_ Echo) GetType() CommandType {
	return CommandEcho
}

func (echoCmd Echo) Handle() (string, error) {
	return types.CreateBulkString(echoCmd.Message), nil
}
