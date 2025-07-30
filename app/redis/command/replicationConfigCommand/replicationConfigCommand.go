package replicationconfig

import (
	"strconv"
	"strings"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ReplicationConfigCommand struct {
	Args []string
}

type ReplicationConfigGetAckCommand struct {
	Args []string
}

func NewReplicationConfigCommand(args []string) (interfaces.Command, error) {
	if len(args) == 0 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	if strings.ToUpper(args[0]) == "GETACK" {
		return &ReplicationConfigGetAckCommand{Args: args}, nil
	}

	return &ReplicationConfigCommand{Args: args}, nil
}

// --------------------- Regular ----------------------

func (cmd *ReplicationConfigCommand) Handle() (string, error) {
	// For now just return OK
	return types.OkString, nil
}

func (cmd *ReplicationConfigCommand) IsHandshakeCommand() bool {
	return true
}

func (cmd *ReplicationConfigCommand) GetHandshakeStep() interfaces.HandshakeStep {
	if len(cmd.Args) == 0 {
		return interfaces.HandshakeStepNone
	}
	if cmd.Args[0] == "listening-port" {
		return interfaces.HandshakeStepReplConfFirst
	} else if cmd.Args[0] == "capa" {
		return interfaces.HandshakeStepReplConfSecond
	} else {
		return interfaces.HandshakeStepNone
	}
}

// --------------------- GetAck ----------------------

func (cmd *ReplicationConfigGetAckCommand) Handle() (string, error) {
	config := redisConfig.NewRedisConfig()
	replicationDetails := config.GetReplicationDetails()
	offset := replicationDetails.ReplicaOffset
	respParts := []string{"REPLCONF", "ACK", strconv.Itoa(offset)}
	return types.CreateBulkStringArray(respParts), nil
}

func (cmd *ReplicationConfigGetAckCommand) IsMasterResponseCommand() bool {
	return true
}
