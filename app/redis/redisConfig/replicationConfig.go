package redisConfig

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"time"
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

func (rd *ReplicationDetails) SendHandshake() (bool, error) {
	if rd.Role == RoleMaster {
		return false, fmt.Errorf("master node does not send handshake")
	}
	if rd.MasterDetails == nil {
		return false, fmt.Errorf("master details not set for slave node")
	}
	if err := rd.sendPing(); err != nil {
		return false, fmt.Errorf("failed to send PING to master: %w", err)
	}
	return true, nil
}

func (rd *ReplicationDetails) sendPing() error {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	address := fmt.Sprintf("%s:%d", rd.MasterDetails.Host, rd.MasterDetails.Port)
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to master %s: %w", address, err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		return fmt.Errorf("failed to send PING command to master %s: %w", address, err)
	}
	return nil
}
