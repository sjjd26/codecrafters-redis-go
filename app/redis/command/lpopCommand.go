package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type LPopCommand struct {
	Key   string
	Store *store.RedisStore
}

func NewLPopCommand(args []string, ctx *CommandContext) (Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	} else if len(args) > 1 {
		return nil, cmderrors.ErrTooManyArgs
	}

	key := args[0]
	return &LPopCommand{
		Key:   key,
		Store: ctx.Store,
	}, nil
}

func (cmd *LPopCommand) Execute() (string, error) {
	list, exists := cmd.Store.GetList(cmd.Key)
	if !exists {
		return types.BulkStringNull, nil
	}
	value := list[0]
	list = list[1:]
	listValue := store.NewRedisList(list)
	cmd.Store.Add(cmd.Key, listValue)
	return types.CreateBulkString(value), nil
}
