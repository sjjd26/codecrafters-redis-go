package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type Echo struct {
	Message string
}

func NewEcho(args []string) (interfaces.Command, error) {
	if len(args) == 0 {
		return nil, cmderrors.ErrNotEnoughArgs
	} else if len(args) > 1 {
		return nil, cmderrors.ErrTooManyArgs
	}

	return Echo{Message: args[0]}, nil
}

func (_ Echo) GetType() CommandType {
	return CommandEcho
}

func (echoCmd Echo) Handle() (string, error) {
	return types.CreateBulkString(echoCmd.Message), nil
}
