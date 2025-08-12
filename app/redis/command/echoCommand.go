package command

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type EchoCommand struct {
	Message string
}

func NewEchoCommand(args []string, _ *CommandContext) (Command, error) {
	if len(args) == 0 {
		return nil, cmderrors.ErrNotEnoughArgs
	} else if len(args) > 1 {
		return nil, cmderrors.ErrTooManyArgs
	}

	return EchoCommand{Message: args[0]}, nil
}

func (echoCmd EchoCommand) Execute() (string, error) {
	return types.CreateBulkString(echoCmd.Message), nil
}
