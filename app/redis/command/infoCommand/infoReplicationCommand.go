package infocommand

import (
	"fmt"
	"strings"

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
	masterReplId := fmt.Sprintf("master_replid:%s", replicationDetails.MasterReplId)
	masterReplOffset := fmt.Sprintf("master_repl_offset:%d", replicationDetails.MasterReplOffset)
	return types.CreateBulkString(strings.Join([]string{role, masterReplId, masterReplOffset}, "\n")), nil
}
