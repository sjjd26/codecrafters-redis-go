package redisConfig

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
	Role          RedisInfoRole
	SelfDetails   *HostDetails
	MasterDetails *HostDetails
}
