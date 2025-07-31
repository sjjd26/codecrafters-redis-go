package command

import (
	"strings"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ConfigGetCommand struct {
	Key    redisConfig.ConfigKey
	Config redisConfig.RedisConfig
}

func NewConfigCommand(args []string, ctx *CommandContext) (interfaces.Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	subCommand := strings.ToUpper(args[0])
	switch subCommand {
	case "SET":
		return nil, cmderrors.ErrNotImplemented("CONFIG SET")
	case "GET":
		return NewConfigGetCommand(args[1:], ctx)
	default:
		return nil, cmderrors.ErrUnknownSubCommand
	}
}

func NewConfigGetCommand(args []string, ctx *CommandContext) (interfaces.Command, error) {
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
	return &ConfigGetCommand{
		Key:    key,
		Config: ctx.Config,
	}, nil
}

func (cmd *ConfigGetCommand) Handle() (string, error) {
	value, ok := cmd.Config.Get(cmd.Key)
	if ok {
		return types.CreateKeyValueArray(string(cmd.Key), value), nil
	}
	return types.CreateKeyValueNullArray(string(cmd.Key)), nil
}
