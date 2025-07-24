package redisConfig

var redisConfig RedisConfig = nil

type RedisConfig interface {
	Get(ConfigKey) (string, bool)
	Set(ConfigKey, string)
	SetReplicationDetails(*ReplicationDetails)
	GetReplicationDetails() *ReplicationDetails
}

type RedisConfigImpl struct {
	ConfigStore        map[ConfigKey]string
	ReplicationDetails *ReplicationDetails
}

func NewRedisConfig() RedisConfig {
	if redisConfig == nil {
		redisConfig = &RedisConfigImpl{
			ConfigStore:        make(map[ConfigKey]string),
			ReplicationDetails: nil,
		}
	}
	return redisConfig
}

func (rc *RedisConfigImpl) Get(key ConfigKey) (string, bool) {
	if value, ok := rc.ConfigStore[key]; ok {
		return value, true
	}
	return "", false
}

func (rc *RedisConfigImpl) Set(key ConfigKey, value string) {
	rc.ConfigStore[key] = value
}

func (rc *RedisConfigImpl) GetReplicationDetails() *ReplicationDetails {
	return rc.ReplicationDetails
}

func (rc *RedisConfigImpl) SetReplicationDetails(details *ReplicationDetails) {
	rc.ReplicationDetails = details
}
