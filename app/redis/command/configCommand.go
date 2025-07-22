package command

import (
	"strings"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ConfigGet struct {
	key redisConfig.ConfigKey
}

func NewConfig(args []string) (interfaces.Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	subCommand := strings.ToUpper(args[0])
	switch subCommand {
	case "SET":
		return nil, cmderrors.ErrNotImplemented("CONFIG SET")
	case "GET":
		return NewGetCommand(args[1:])
	default:
		return nil, cmderrors.ErrUnknownSubCommand
	}
}

func NewGetCommand(args []string) (interfaces.Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}
	if len(args) > 1 {
		return nil, cmderrors.ErrTooManyArgs
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
