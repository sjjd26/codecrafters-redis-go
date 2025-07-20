package command

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ConfigGet struct {
	key redisConfig.ConfigKey
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

	key, err := redisConfig.NewConfigKey(args[0])
	if err != nil {
		return nil, err
	}
	return ConfigGet{key: key}, nil
}

func (cmd ConfigGet) Handle() (string, error) {
	rConfig := redisConfig.NewRedisConfig()
	value, ok := rConfig.Get(cmd.key)
	if ok {
		return types.CreateKeyValueArray(string(cmd.key), value), nil
	}
	return types.CreateKeyValueNullArray(string(cmd.key)), nil
}
