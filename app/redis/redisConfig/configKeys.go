package redisConfig

import "fmt"

type ConfigKey string

const (
	ConfigUnknown    ConfigKey = "unknown"
	ConfigDir        ConfigKey = "dir"
	ConfigDbFilename ConfigKey = "dbfilename"
)

var configKeyMap = map[string]ConfigKey{
	"dir":        ConfigDir,
	"dbfilename": ConfigDbFilename,
}

func NewConfigKey(key string) (ConfigKey, error) {
	configKey, ok := configKeyMap[key]
	if !ok {
		return ConfigUnknown, fmt.Errorf("unknown config key: %s", key)
	}
	return configKey, nil
}
