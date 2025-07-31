package command

import (
	"fmt"
	"strconv"

	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/rdbRestorer"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
)

type PsyncCommand struct {
	ReplicationId      string
	ReplicationOffset  int
	ReplicationDetails *redisConfig.ReplicationDetails
	Store              store.RedisStore
}

func NewPsyncCommand(args []string, ctx *CommandContext) (interfaces.Command, error) {
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
		ReplicationId:      args[0],
		ReplicationOffset:  offset,
		ReplicationDetails: ctx.ReplicationDetails,
		Store:              ctx.Store,
	}, nil
}

func (cmd *PsyncCommand) Execute() (string, error) {
	if cmd.ReplicationId != "?" || cmd.ReplicationOffset != -1 {
		return "", fmt.Errorf("replication and/or offset values not supported: %s, %d", cmd.ReplicationId, cmd.ReplicationOffset)
	}

	fullResync := fmt.Sprintf("+FULLRESYNC %s %s\r\n", cmd.ReplicationDetails.MasterReplId, strconv.Itoa(cmd.ReplicationDetails.MasterReplOffset))
	rdbData, err := cmd.getRdbFileData()
	if err != nil {
		return "", fmt.Errorf("failed to get RDB file data: %w", err)
	}

	response := fullResync + rdbData
	return response, nil
}

func (cmd *PsyncCommand) getRdbFileData() (string, error) {
	restorer, err := rdbRestorer.NewRdbRestorer(cmd.Store)
	if err != nil {
		return "", fmt.Errorf("failed to create restorer: %w", err)
	}
	rdbData, err := restorer.SaveStoreToRdb()
	if err != nil {
		return "", fmt.Errorf("failed to restore to RDB: %w", err)
	}

	rdbDataLen := len(rdbData)
	rdbDataString := fmt.Sprintf("$%d\r\n%s", rdbDataLen, rdbData)

	return rdbDataString, nil
}

func (cmd *PsyncCommand) IsHandshakeCommand() bool {
	return true
}

func (cmd *PsyncCommand) GetHandshakeStep() interfaces.HandshakeStep {
	return interfaces.HandshakeStepPsync
}
