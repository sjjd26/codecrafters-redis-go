package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type KeysCommand struct {
	pattern string
}

func NewKeysCommand(args []string) (interfaces.Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}
	if len(args) > 1 {
		return nil, cmderrors.ErrTooManyArgs
	}

	return &KeysCommand{pattern: args[0]}, nil
}

func (cmd *KeysCommand) Handle() (string, error) {
	redisStore := store.NewRedisStore()
	keys, err := redisStore.GetKeysByPattern(cmd.pattern)
	if err != nil {
		return "", err
	}

	if len(keys) == 0 {
		return types.CreateEmptyArray(), nil
	}

	return types.CreateBulkStringArray(keys), nil
}
