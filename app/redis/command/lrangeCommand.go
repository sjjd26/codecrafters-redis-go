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

	// fmt.Println("LRangeCommand: key =", key, ", start =", start, ", end =", end)

	return &LRangeCommand{
		Key:   key,
		Start: start,
		End:   end,
		Store: ctx.Store,
	}, nil
}

func (cmd *LRangeCommand) Execute() (string, error) {
	list, exists := cmd.Store.GetList(cmd.Key)
	start := cmd.Start
	end := cmd.End
	if !exists {
		return types.CreateEmptyArray(), nil
	}
	if start < 0 {
		start = int(math.Max(float64(len(list)+start), 0))
	}
	if end < 0 {
		end = int(math.Max(float64(len(list)+end), 0))
	}
	if start > end {
		return types.CreateEmptyArray(), nil
	}
	end = int(math.Min(float64(len(list)-1), float64(end)))
	// fmt.Println("start =", start, ", end =", end, ", list length =", len(list))
	return types.CreateBulkStringArray(list[start:int(end+1)]), nil
}
