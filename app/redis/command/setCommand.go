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

type Set struct {
	key    string
	value  string
	expiry int64
}

func NewSet(args []string) (interfaces.Command, error) {
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

	return Set{key: key, value: value, expiry: expiry}, nil
}

func (_ Set) GetType() CommandType {
	return CommandSet
}

func (cmd Set) Handle() (string, error) {
	rStore := store.NewRedisStore()
	rStore.Add(cmd.key, cmd.value)
	if cmd.expiry > 0 {
		now := time.Now().UnixMilli()
		rStore.AddExpiry(cmd.key, cmd.expiry+now)
	}
	return types.OkString, nil
}

func (cmd Set) IsWriteCommand() bool {
	return true
}
