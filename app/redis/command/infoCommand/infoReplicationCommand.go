package infocommand

import (
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ReplicationCommand struct{}

func NewReplicationCommand(args []string) (interfaces.Command, error) {
	return &ReplicationCommand{}, nil
}

func (cmd *ReplicationCommand) Handle() (string, error) {
	role := "role:master"
	return types.CreateBulkString(role), nil
}
