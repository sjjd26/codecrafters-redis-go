package store

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/gobwas/glob"
)

var redisStore RedisStore = nil

type RedisStore interface {
	Add(key string, value string) bool
	AddExpiry(key string, expiry int64) error
	Get(key string) (string, bool)
	GetKeysByPattern(pattern string) ([]string, error)
}

type RedisStoreImpl struct {
	valueStore  map[string]string
	expiryStore map[string]int64
	config      redisConfig.RedisConfig
}

func NewRedisStore() RedisStore {
	if redisStore == nil {
		redisStore = &RedisStoreImpl{
			valueStore:  make(map[string]string),
			expiryStore: make(map[string]int64),
			config:      redisConfig.NewRedisConfig(),
		}
	}
	return redisStore
}

// Add adds a key-value pair to the store.
// Returns true if the key already existed, false otherwise.
func (rs *RedisStoreImpl) Add(key string, value string) bool {
	_, exists := rs.valueStore[key]
	rs.valueStore[key] = value
	return exists
}

func (rs *RedisStoreImpl) AddExpiry(key string, expiry int64) error {
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

func (rs *RedisStoreImpl) Get(key string) (string, bool) {
	value, ok := rs.valueStore[key]
	if !ok {
		return "", false
	}

	var expiry int64
	if expiry, ok = rs.expiryStore[key]; !ok || expiry == 0 {
		return value, true
	}

	now := time.Now().UnixMilli()
	if now > expiry {
		delete(rs.valueStore, key)
		delete(rs.expiryStore, key)
		return "", false
	}

	return value, true
}

func (rs *RedisStoreImpl) GetKeysByPattern(pattern string) ([]string, error) {
	var keys []string
	glob := glob.MustCompile(pattern)
	for key := range rs.valueStore {
		if glob.Match(key) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}
