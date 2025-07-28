package replicationconfig

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ReplicationConfigCommand struct {
	Key   string
	Value string
}

func NewReplicationConfigCommand(args []string) (interfaces.Command, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("invalid number of arguments for REPLCONF command")
	}

	return &ReplicationConfigCommand{
		Key:   args[0],
		Value: args[1],
	}, nil
}

func (cmd *ReplicationConfigCommand) Handle() (string, error) {
	// For now just return OK
	return types.OkString, nil
}
