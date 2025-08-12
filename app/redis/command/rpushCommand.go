package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type RPushCommand struct {
	Key    string
	Values []string
	Store  *store.RedisStore
}

func NewRPushCommand(args []string, ctx *CommandContext) (Command, error) {
	if len(args) < 2 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	key := args[0]
	values := args[1:]

	return &RPushCommand{
		Key:    key,
		Values: values,
		Store:  ctx.Store,
	}, nil
}

func (cmd *RPushCommand) Execute() (string, error) {
	list, exists := cmd.Store.GetList(cmd.Key)
	if !exists {
		list = []string{}
	}
	list = append(list, cmd.Values...)
	listValue := store.NewRedisList(list)
	cmd.Store.Add(cmd.Key, listValue)
	return types.CreateInt(len(list)), nil
}

func (cmd *RPushCommand) IsWriteCommand() bool {
	return true
}
