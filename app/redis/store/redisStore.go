package store

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/gobwas/glob"
)

var redisStore *RedisStore = nil

// type RedisStore interface {
// 	Add(key string, value RedisValue) bool
// 	AddExpiry(key string, expiry int64) error
// 	GetString(key string) (string, bool)
// 	GetKeysByPattern(pattern string) ([]string, error)
// 	Size() int
// }

type RedisStore struct {
	valueStore  map[string]RedisValue
	expiryStore map[string]int64
	config      redisConfig.RedisConfig
}

func NewRedisStore() *RedisStore {
	if redisStore == nil {
		redisStore = &RedisStore{
			valueStore:  make(map[string]RedisValue),
			expiryStore: make(map[string]int64),
			config:      redisConfig.NewRedisConfig(),
		}
	}
	return redisStore
}

// Add adds a key-value pair to the store.
// Returns true if the key already existed, false otherwise.
func (rs *RedisStore) Add(key string, value RedisValue) bool {
	_, exists := rs.valueStore[key]
	rs.valueStore[key] = value
	return exists
}

func (rs *RedisStore) AddExpiry(key string, expiry int64) error {
	if _, exists := rs.valueStore[key]; !exists {
		return fmt.Errorf("key %s does not exist", key)
	}
	if expiry == 0 {
		delete(rs.expiryStore, key)
		return nil
	}
	rs.expiryStore[key] = expiry
	return nil
}

func (rs *RedisStore) GetString(key string) (string, bool) {
	value, ok := rs.valueStore[key]
	if !ok {
		return "", false
	}

	var expiry int64
	if expiry, ok = rs.expiryStore[key]; !ok || expiry == 0 {
		stringValue, err := ExtractString(value)
		if err != nil {
			return "", false
		}
		return stringValue, true
	}

	now := time.Now().UnixMilli()
	if now > expiry {
		delete(rs.valueStore, key)
		delete(rs.expiryStore, key)
		return "", false
	}

	stringValue, err := ExtractString(value)
	if err != nil {
		return "", false
	}
	return stringValue, true
}

func (rs *RedisStore) GetList(key string) ([]string, bool) {
	value, ok := rs.valueStore[key]
	if !ok {
		return nil, false
	}

	var expiry int64
	if expiry, ok = rs.expiryStore[key]; !ok || expiry == 0 {
		listValue, err := ExtractList(value)
		if err != nil {
			return nil, false
		}
		return listValue, true
	}

	now := time.Now().UnixMilli()
	if now > expiry {
		delete(rs.valueStore, key)
		delete(rs.expiryStore, key)
		return nil, false
	}

	listValue, err := ExtractList(value)
	if err != nil {
		return nil, false
	}
	return listValue, true
}

func (rs *RedisStore) GetKeysByPattern(pattern string) ([]string, error) {
	var keys []string
	glob := glob.MustCompile(pattern)
	for key := range rs.valueStore {
		if glob.Match(key) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (rs *RedisStore) Size() int {
	return len(rs.valueStore)
}
