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

func NewReplicationDetails(role RedisInfoRole, selfPort int) (*ReplicationDetails, error) {
	id, err := generateRandomId(40)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random ID for replication details: %w", err)
	}
	return &ReplicationDetails{
		Role:             role,
		MasterReplId:     id,
		MasterReplOffset: 0,
		SelfDetails: &HostDetails{
			Host: "localhost",
			Port: selfPort,
		},
	}, nil
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
