package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type PingCommand struct {
	message string
}

func NewPing(args []string) (interfaces.Command, error) {
	if len(args) > 1 {
		return nil, cmderrors.ErrTooManyArgs
	}

	if len(args) > 0 {
		return &PingCommand{message: args[0]}, nil
	}
	return &PingCommand{}, nil
}

func (_ *PingCommand) GetType() CommandType {
	return CommandPing
}

func (cmd *PingCommand) Handle() (string, error) {
	if cmd.message != "" {
		return types.CreateBulkString(cmd.message), nil
	}
	return "+PONG\r\n", nil
}

func (cmd *PingCommand) IsHandshakeCommand() bool {
	return true
}

func (cmd *PingCommand) GetHandshakeStep() interfaces.HandshakeStep {
	return interfaces.HandshakeStepPing
}
