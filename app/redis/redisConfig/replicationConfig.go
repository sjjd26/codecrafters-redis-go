package redisConfig

import (
	"crypto/rand"
	"fmt"
)

type HostDetails struct {
	Host string
	Port int
}

type RedisInfoRole string

const (
	RoleMaster RedisInfoRole = "master"
	RoleSlave  RedisInfoRole = "slave"
)

type ReplicationDetails struct {
	Role             RedisInfoRole
	SelfDetails      *HostDetails
	MasterDetails    *HostDetails
	MasterReplId     string
	MasterReplOffset int
}

func NewReplicationDetails(selfDetails, masterDetails *HostDetails) (*ReplicationDetails, error) {
	var err error
	var id string = "?"
	masterReplOffset := -1
	role := RoleSlave

	if selfDetails == nil {
		return nil, fmt.Errorf("self details cannot be nil")
	}

	if masterDetails == nil {
		role = RoleMaster
		masterReplOffset = 0
		id, err = generateRandomId(40)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random ID for replication details: %w", err)
		}
	}

	replicationDetails := &ReplicationDetails{
		Role:             role,
		MasterReplId:     id,
		MasterReplOffset: masterReplOffset,
		SelfDetails:      selfDetails,
		MasterDetails:    masterDetails,
	}

	return replicationDetails, nil
}

func generateRandomId(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const charsetLen = len(charset)
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		b[i] = charset[int(b[i])%charsetLen]
	}
	return string(b), nil
}
