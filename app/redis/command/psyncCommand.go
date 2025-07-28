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

	fullResync := fmt.Sprintf("+FULLRESYNC %s %s\r\n", replDetails.MasterReplId, strconv.Itoa(replDetails.MasterReplOffset))
	rdbData, err := cmd.getRdbFileData()
	if err != nil {
		return "", fmt.Errorf("failed to get RDB file data: %w", err)
	}

	response := fullResync + rdbData
	return response, nil
}

func (cmd *PsyncCommand) getRdbFileData() (string, error) {
	redisStore := store.NewRedisStore()
	restorer := rdbRestorer.NewRdbRestorer(redisStore)
	rdbData, err := restorer.SaveStoreToRdb()
	if err != nil {
		return "", fmt.Errorf("failed to restore to RDB: %w", err)
	}

	rdbDataLen := len(rdbData)
	rdbDataString := fmt.Sprintf("$%d\r\n%s", rdbDataLen, rdbData)

	return rdbDataString, nil
}
