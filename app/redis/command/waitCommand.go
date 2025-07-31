package command

import (
	"fmt"
	"strconv"
	"time"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type WaitCommand struct {
	ReplicaCount       int
	Timeout            int
	Start              int64
	ReplicationDetails *redisConfig.ReplicationDetails
}

func NewWaitCommand(args []string, ctx *CommandContext) (interfaces.Command, error) {
	if len(args) < 2 {
		return nil, cmderrors.ErrNotEnoughArgs
	}
	if len(args) > 2 {
		return nil, cmderrors.ErrTooManyArgs
	}

	replicaCount, err := strconv.Atoi(args[0])
	if err != nil || replicaCount < 0 {
		return nil, fmt.Errorf("invalid replica count: %w", err)
	}

	timeout, err := strconv.Atoi(args[1])
	if err != nil || timeout < 0 {
		return nil, fmt.Errorf("invalid timeout value: %w", err)
	}
	start := time.Now().UnixMilli()

	return &WaitCommand{
		ReplicaCount:       replicaCount,
		Timeout:            timeout,
		Start:              start,
		ReplicationDetails: ctx.ReplicationDetails,
	}, nil
}

func (cmd *WaitCommand) Execute() (string, error) {
	if cmd.ReplicationDetails.Role != redisConfig.RoleMaster {
		return "", fmt.Errorf("WAIT command can only be executed on master nodes")
	}

	replCount := len(cmd.ReplicationDetails.SlaveConnections)
	if cmd.ReplicaCount == 0 {
		return types.CreateInt(replCount), nil
	}

	timeDiff := time.Now().UnixMilli() - cmd.Start
	if cmd.Timeout > 0 && timeDiff < int64(cmd.Timeout) {
		time.Sleep(time.Duration(timeDiff) * time.Millisecond)
	}

	return types.CreateInt(replCount), nil
}
