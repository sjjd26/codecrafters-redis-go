package command

import (
	"fmt"
	"strconv"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type LPopCommand struct {
	Key           string
	CountProvided bool
	Count         int
	Store         *store.RedisStore
}

func NewLPopCommand(args []string, ctx *CommandContext) (Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	} else if len(args) > 2 {
		return nil, cmderrors.ErrTooManyArgs
	}

	key := args[0]
	count := 1
	countProvided := false
	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("invalid count argument: %v", err)
		}
		countProvided = true
	}

	return &LPopCommand{
		Key:           key,
		Count:         count,
		CountProvided: countProvided,
		Store:         ctx.Store,
	}, nil
}

func (cmd *LPopCommand) Execute() (string, error) {
	list, exists := cmd.Store.GetList(cmd.Key)
	if !exists || len(list) == 0 {
		return types.BulkStringNull, nil
	}

	if cmd.Count > len(list) {
		cmd.Count = len(list)
	}
	values := list[:cmd.Count]
	list = list[cmd.Count:]
	listValue := store.NewRedisList(list)
	cmd.Store.Add(cmd.Key, listValue)

	if cmd.CountProvided {
		return types.CreateBulkStringArray(values), nil
	}
	return types.CreateBulkString(values[0]), nil
}
