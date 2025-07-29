package replicationconfig

import (
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type ReplicationConfigCommand struct {
	Args []string
}

func NewReplicationConfigCommand(args []string) (interfaces.Command, error) {
	return &ReplicationConfigCommand{Args: args}, nil
}

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
