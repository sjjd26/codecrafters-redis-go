package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type KeysCommand struct {
	Pattern string
	Store   *store.RedisStore
}

func NewKeysCommand(args []string, ctx *CommandContext) (Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}
	if len(args) > 1 {
		return nil, cmderrors.ErrTooManyArgs
	}

	return &KeysCommand{
		Pattern: args[0],
		Store:   ctx.Store,
	}, nil
}

func (cmd *KeysCommand) Execute() (string, error) {
	keys, err := cmd.Store.GetKeysByPattern(cmd.Pattern)
	if err != nil {
		return "", err
	}

	if len(keys) == 0 {
		return types.CreateEmptyArray(), nil
	}

	return types.CreateBulkStringArray(keys), nil
}
