package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type PingCommand struct {
	Message string
}

func NewPingCommand(args []string, _ *CommandContext) (interfaces.Command, error) {
	if len(args) > 1 {
		return nil, cmderrors.ErrTooManyArgs
	}

	if len(args) > 0 {
		return &PingCommand{Message: args[0]}, nil
	}
	return &PingCommand{}, nil
}

func (cmd *PingCommand) Handle() (string, error) {
	if cmd.Message != "" {
		return types.CreateBulkString(cmd.Message), nil
	}
	return "+PONG\r\n", nil
}

func (cmd *PingCommand) IsHandshakeCommand() bool {
	return true
}

func (cmd *PingCommand) GetHandshakeStep() interfaces.HandshakeStep {
	return interfaces.HandshakeStepPing
}
