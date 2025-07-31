package command

import (
	"fmt"
	"strings"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type InfoSubCommand string

const (
	InfoSubCommandUnknown     InfoSubCommand = "unknown"
	InfoSubCommandReplication InfoSubCommand = "replication"
)

var InfoSubCommandMap = map[string]InfoSubCommand{
	"replication": InfoSubCommandReplication,
}

func NewInfoCommand(args []string, ctx *CommandContext) (interfaces.Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	subCommand := InfoSubCommandMap[args[0]]
	switch subCommand {
	case InfoSubCommandReplication:
		return NewReplicationCommand(args[1:], ctx)
	default:
		return nil, cmderrors.ErrUnknownSubCommand
	}
}

// -------------------- Replication Command --------------------

type ReplicationCommand struct{}

func NewReplicationCommand(args []string, _ *CommandContext) (interfaces.Command, error) {
	return &ReplicationCommand{}, nil
}

func (cmd *ReplicationCommand) Execute() (string, error) {
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
