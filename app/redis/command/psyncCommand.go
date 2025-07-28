package command

import (
	"fmt"
	"strconv"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
)

type PsyncCommand struct {
	replicationId     string
	replicationOffset int
}

func NewPsyncCommand(args []string) (interfaces.Command, error) {
	if len(args) < 2 {
		return nil, cmderrors.ErrNotEnoughArgs
	}
	if len(args) > 2 {
		return nil, cmderrors.ErrTooManyArgs
	}

	offset, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse replication offset: %w", err)
	}

	return &PsyncCommand{
		replicationId:     args[0],
		replicationOffset: offset,
	}, nil
}

func (cmd *PsyncCommand) Handle() (string, error) {
	if cmd.replicationId != "?" || cmd.replicationOffset != -1 {
		return "", fmt.Errorf("replication and/or offset values not supported: %s, %d", cmd.replicationId, cmd.replicationOffset)
	}

	config := redisConfig.NewRedisConfig()
	replDetails := config.GetReplicationDetails()

	response := fmt.Sprintf("+FULLRESYNC %s %s\r\n", replDetails.MasterReplId, strconv.Itoa(replDetails.MasterReplOffset))
	return response, nil
}
