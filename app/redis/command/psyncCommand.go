package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type PsyncCommand struct{}

func NewPsyncCommand(args []string) (interfaces.Command, error) {
	return &PsyncCommand{}, nil
}

func (cmd *PsyncCommand) Handle() (string, error) {
	return types.OkString, nil
}
