package redisConfig

var redisConfig RedisConfig = nil

type RedisConfig interface {
	Get(ConfigKey) (string, bool)
	Set(ConfigKey, string)
}

type RedisConfigImpl struct {
	configStore map[ConfigKey]string
}

func NewRedisConfig() RedisConfig {
	if redisConfig == nil {
		redisConfig = RedisConfigImpl{
			configStore: make(map[ConfigKey]string),
		}
	}
	return redisConfig
}

func (rc RedisConfigImpl) Get(key ConfigKey) (string, bool) {
	if value, ok := rc.configStore[key]; ok {
		return value, true
	}
	return "", false
}

func (rc RedisConfigImpl) Set(key ConfigKey, value string) {
	rc.configStore[key] = value
}
