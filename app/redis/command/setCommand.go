package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type SetCommand struct {
	Key    string
	Value  string
	Expiry int64
	Store  store.RedisStore
}

func NewSetCommand(args []string, ctx *CommandContext) (interfaces.Command, error) {
	if len(args) < 2 {
		return nil, cmderrors.ErrNotEnoughArgs
	}
	key := args[0]
	value := args[1]
	expiry := int64(0)

	if len(args) > 3 {
		isPx := strings.ToUpper(args[2]) == "PX"
		if !isPx {
			return nil, fmt.Errorf("argument type %v not implemented", args[2])
		}
		var err error
		expiry, err = strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse expiry value %v: %w", args[3], err)
		}
	}

	return SetCommand{
		Key:    key,
		Value:  value,
		Expiry: expiry,
		Store:  ctx.Store,
	}, nil
}

func (cmd SetCommand) Handle() (string, error) {
	cmd.Store.Add(cmd.Key, cmd.Value)
	if cmd.Expiry > 0 {
		now := time.Now().UnixMilli()
		cmd.Store.AddExpiry(cmd.Key, cmd.Expiry+now)
	}
	return types.OkString, nil
}

func (cmd SetCommand) IsWriteCommand() bool {
	return true
}
