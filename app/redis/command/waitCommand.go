package command

import (
	"fmt"
	"strconv"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type WaitCommand struct {
	ReplicaCount int
	Timeout      int
}

func NewWaitCommand(args []string) (interfaces.Command, error) {
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

	return &WaitCommand{
		ReplicaCount: replicaCount,
		Timeout:      timeout,
	}, nil
}

func (cmd *WaitCommand) Handle() (string, error) {
	// For now, just return 0
	return types.CreateInt(0), nil
}
