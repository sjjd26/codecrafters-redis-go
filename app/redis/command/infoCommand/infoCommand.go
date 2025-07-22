package infocommand

import (
	cmderrors "github.com/codecrafters-io/redis-starter-go/app/redis/command/cmdErrors"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
)

type InfoSubCommand string

const (
	InfoSubCommandUnknown     InfoSubCommand = "unknown"
	InfoSubCommandReplication InfoSubCommand = "replication"
)

var InfoSubCommandMap = map[string]InfoSubCommand{
	"replication": InfoSubCommandReplication,
}

func NewInfoCommand(args []string) (interfaces.Command, error) {
	if len(args) < 1 {
		return nil, cmderrors.ErrNotEnoughArgs
	}

	subCommand := InfoSubCommandMap[args[0]]
	switch subCommand {
	case InfoSubCommandReplication:
		return NewReplicationCommand(args[1:])
	default:
		return nil, cmderrors.ErrUnknownSubCommand
	}
}
