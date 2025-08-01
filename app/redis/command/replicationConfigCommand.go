package command

import (
	"fmt"
	"net"
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

func NewReplicationConfigCommand(args []string, ctx *CommandContext) (interfaces.Command, error) {
	if len(args) == 0 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	if strings.ToUpper(args[0]) == "GETACK" {
		return &ReplicationConfigGetAckCommand{
			ReplicationDetails: ctx.ReplicationDetails,
		}, nil
	}

	if strings.ToUpper(args[0]) == "ACK" {
		if len(args) < 2 {
			return nil, cmderrors.ErrNotEnoughArgs
		}
		offset, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("invalid offset value: %s", args[1])
		}
		return &ReplicationConfigAckCommand{
			Offset:             offset,
			ReplicationDetails: ctx.ReplicationDetails,
			Conn:               ctx.Conn,
		}, nil
	}

	return &ReplicationConfigCommand{Args: args}, nil
}

// --------------------- Regular ----------------------

func (cmd *ReplicationConfigCommand) Execute() (string, error) {
	// For now just return OK
	return types.OkString, nil
}

func (cmd *ReplicationConfigCommand) IsHandshakeCommand() bool {
	return true
}

func (cmd *ReplicationConfigCommand) GetHandshakeStep() interfaces.HandshakeStep {
	if cmd.Args[0] == "listening-port" {
		return interfaces.HandshakeStepReplConfFirst
	} else if cmd.Args[0] == "capa" {
		return interfaces.HandshakeStepReplConfSecond
	} else {
		return interfaces.HandshakeStepNone
	}
}

// --------------------- GetAck ----------------------

type ReplicationConfigGetAckCommand struct {
	ReplicationDetails *redisConfig.ReplicationDetails
}

func (cmd *ReplicationConfigGetAckCommand) Execute() (string, error) {
	offset := cmd.ReplicationDetails.ReplicaOffset
	respParts := []string{"REPLCONF", "ACK", strconv.Itoa(offset)}
	return types.CreateBulkStringArray(respParts), nil
}

func (cmd *ReplicationConfigGetAckCommand) IsMasterResponseCommand() bool {
	return true
}

// -------------------- Ack ----------------------

type ReplicationConfigAckCommand struct {
	Offset             int
	ReplicationDetails *redisConfig.ReplicationDetails
	Conn               net.Conn
}

func (cmd *ReplicationConfigAckCommand) Execute() (string, error) {
	replicaDetails, ok := cmd.ReplicationDetails.SlaveConnections[cmd.Conn]
	if !ok {
		return "", fmt.Errorf("replica connection not found in replication details")
	}
	replicaDetails.LatestOffset = cmd.Offset
	return "", nil
}
