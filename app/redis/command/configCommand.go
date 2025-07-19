package command

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ConfigGet struct {
	key string
}

func NewConfig(args []string) (Command, error) {
	if len(args) < 1 {
		return nil, ErrNotEnoughArgs
	}

	subCommand := strings.ToUpper(args[0])
	switch subCommand {
	case "SET":
		return nil, ErrNotImplemented("CONFIG SET")
	case "GET":
		return NewGetCommand(args[1:])
	default:
		return nil, ErrUnknownSubCommand
	}
}

func NewGetCommand(args []string) (Command, error) {
	if len(args) < 1 {
		return nil, ErrNotEnoughArgs
	}
	if len(args) > 1 {
		return nil, ErrTooManyArgs
	}

	key := args[0]
	return ConfigGet{key: key}, nil
}

// func (c ConfigGet) GetType() CommandType {
// 	return CommandConfigGet
// }

func (cmd ConfigGet) Handle() (string, error) {
	value, ok := store.GetConfig(cmd.key)
	if ok {
		return types.CreateKeyValueArray(cmd.key, value), nil
	}
	return types.CreateKeyValueNullArray(cmd.key), nil
}
