package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type GetCommand struct {
	Key   string
	Store store.RedisStore
}

func NewGetCommand(args []string, ctx *CommandContext) (interfaces.Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	key := args[0]
	return &GetCommand{
		Key:   key,
		Store: ctx.Store,
	}, nil
}

func (cmd *GetCommand) Handle() (string, error) {
	if value, ok := cmd.Store.Get(cmd.Key); ok {
		return types.CreateBulkString(value), nil
	}
	return types.BulkStringNull, nil
}
