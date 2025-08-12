package command

import (
	"slices"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type LPushCommand struct {
	Key    string
	Values []string
	Store  *store.RedisStore
}

func NewLPushCommand(args []string, ctx *CommandContext) (Command, error) {
	if len(args) < 2 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	key := args[0]
	values := args[1:]
	slices.Reverse(values) // Reverse values to maintain order when prepending

	return &LPushCommand{
		Key:    key,
		Values: values,
		Store:  ctx.Store,
	}, nil
}

func (cmd *LPushCommand) Execute() (string, error) {
	list, exists := cmd.Store.GetList(cmd.Key)
	if !exists {
		list = []string{}
	}
	list = append(cmd.Values, list...)
	listValue := store.NewRedisList(list)
	cmd.Store.Add(cmd.Key, listValue)
	return types.CreateInt(len(list)), nil
}

func (cmd *LPushCommand) IsWriteCommand() bool {
	return true
}
