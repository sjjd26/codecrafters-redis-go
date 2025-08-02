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
	Start              time.Time
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
	start := time.Now()

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

	currentUpToDate := cmd.CheckLatestReplicaOffsets()
	if currentUpToDate >= cmd.ReplicaCount {
		return types.CreateInt(currentUpToDate), nil
	}

	err := cmd.BroadcastGetAck()
	if err != nil {
		return "", fmt.Errorf("failed to broadcast REPLCONF GETACK: %w", err)
	}

	timeoutDuration := time.Duration(cmd.Timeout) * time.Millisecond
	for time.Since(cmd.Start) < timeoutDuration {
		currentUpToDate = cmd.CheckLatestReplicaOffsets()
		if currentUpToDate >= cmd.ReplicaCount {
			return types.CreateInt(currentUpToDate), nil
		}
		time.Sleep(10 * time.Millisecond) // Sleep briefly to avoid busy waiting
	}

	return types.CreateInt(currentUpToDate), nil
}

func (cmd *WaitCommand) CheckLatestReplicaOffsets() int {
	var count int
	for _, replica := range cmd.ReplicationDetails.SlaveConnections {
		if replica.LatestOffset >= cmd.ReplicationDetails.ReplicaOffset {
			count++
		}
	}
	return count
}

func (cmd *WaitCommand) BroadcastGetAck() error {
	commandParts := []string{"REPLCONF", "GETACK", "*"}
	command := types.CreateBulkStringArray(commandParts)
	for replicaConn, _ := range cmd.ReplicationDetails.SlaveConnections {
		if _, err := replicaConn.Write([]byte(command)); err != nil {
			return fmt.Errorf("failed to send REPLCONF GETACK to replica: %w", err)
		}
	}
	return nil
}
