package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type Set struct {
	key    string
	value  string
	expiry int
}

func NewSet(args []string) (Command, error) {
	if len(args) < 2 {
		return nil, ErrNotEnoughArgs
	}
	key := args[0]
	value := args[1]
	expiry := 0

	if len(args) > 3 {
		isPx := strings.ToUpper(args[2]) == "PX"
		if !isPx {
			return nil, fmt.Errorf("argument type %v not implemented", args[2])
		}
		var err error
		expiry, err = strconv.Atoi(args[3])
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
	store.Add(cmd.key, cmd.value, cmd.expiry)
	return types.OkString, nil
}
