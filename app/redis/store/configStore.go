package store

var configStore = make(map[string]string)

func AddConfig(key string, value string) {
	configStore[key] = value
}

func GetConfig(key string) (string, bool) {
	value, ok := configStore[key]
	return value, ok
}
