package command

import (
	"fmt"
	"math"
	"strconv"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type LRangeCommand struct {
	Key   string
	Start int
	End   int
	Store *store.RedisStore
}

func NewLRangeCommand(args []string, ctx *CommandContext) (Command, error) {
	if len(args) < 3 {
		return nil, cmderrors.ErrNotEnoughArgs
	}
	if len(args) > 3 {
		return nil, cmderrors.ErrTooManyArgs
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid start index: %w", err)
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, fmt.Errorf("invalid end index: %w", err)
	}

	fmt.Println("LRangeCommand: key =", key, ", start =", start, ", end =", end)

	return &LRangeCommand{
		Key:   key,
		Start: start,
		End:   end,
		Store: ctx.Store,
	}, nil
}

func (cmd *LRangeCommand) Execute() (string, error) {
	if cmd.Start > cmd.End {
		return types.CreateEmptyArray(), nil
	}
	list, exists := cmd.Store.GetList(cmd.Key)
	if !exists {
		return types.CreateEmptyArray(), nil
	}
	end := math.Min(float64(len(list)-1), float64(cmd.End))
	return types.CreateBulkStringArray(list[cmd.Start:int(end+1)]), nil
}
