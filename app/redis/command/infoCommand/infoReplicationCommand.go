package infocommand

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ReplicationCommand struct{}

func NewReplicationCommand(args []string) (interfaces.Command, error) {
	return &ReplicationCommand{}, nil
}

func (cmd *ReplicationCommand) Handle() (string, error) {
	config := redisConfig.NewRedisConfig()
	replicationDetails := config.GetReplicationDetails()
	if replicationDetails == nil {
		return "", fmt.Errorf("replication details not set")
	}
	role := fmt.Sprintf("role:%s", string(replicationDetails.Role))
	return types.CreateBulkString(role), nil
}
