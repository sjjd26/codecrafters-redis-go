package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type Get struct {
	key string
}

func NewGet(args []string) (Command, error) {
	if len(args) < 1 {
		return nil, ErrNotEnoughArgs
	}

	key := args[0]
	return Get{key: key}, nil
}

func (_ Get) GetType() CommandType {
	return CommandGet
}

func (cmd Get) Handle() (string, error) {
	if value, ok := store.Get(cmd.key); ok {
		return types.CreateBulkString(value), nil
	}
	return types.BulkStringNull, nil
}
