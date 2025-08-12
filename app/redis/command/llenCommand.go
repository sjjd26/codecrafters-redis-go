package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type LLenCommand struct {
	Key   string
	Store *store.RedisStore
}

func NewLLenCommand(args []string, ctx *CommandContext) (Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	key := args[0]
	return &LLenCommand{
		Key:   key,
		Store: ctx.Store,
	}, nil
}

func (cmd *LLenCommand) Execute() (string, error) {
	list, exists := cmd.Store.GetList(cmd.Key)
	if !exists {
		return types.CreateInt(0), nil
	}
	return types.CreateInt(len(list)), nil
}
