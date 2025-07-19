package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type Ping struct {
	message string
}

func NewPing(args []string) (Command, error) {
	if len(args) > 1 {
		return nil, ErrTooManyArgs
	}

	if len(args) > 0 {
		return Ping{message: args[0]}, nil
	}
	return Ping{}, nil
}

func (_ Ping) GetType() CommandType {
	return CommandPing
}

func (pingCmd Ping) Handle() (string, error) {
	if pingCmd.message != "" {
		return types.CreateBulkString(pingCmd.message), nil
	}
	return "+PONG\r\n", nil
}
