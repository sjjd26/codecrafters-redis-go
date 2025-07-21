package store

import (
	"fmt"
	"os"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/gobwas/glob"
	"github.com/hdt3213/rdb/parser"
)

var redisStore RedisStore = nil

type RedisStore interface {
	Add(key string, value string) bool
	AddExpiry(key string, expiry int64) error
	Get(key string) (string, bool)
	RdbRestore() error
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
func (rs RedisStoreImpl) Add(key string, value string) bool {
	_, exists := rs.valueStore[key]
	rs.valueStore[key] = value
	return exists
}

func (rs RedisStoreImpl) AddExpiry(key string, expiry int64) error {
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

func (rs RedisStoreImpl) Get(key string) (string, bool) {
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

func (rs RedisStoreImpl) GetKeysByPattern(pattern string) ([]string, error) {
	var keys []string
	glob := glob.MustCompile(pattern)
	for key := range rs.valueStore {
		if glob.Match(key) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (rs RedisStoreImpl) RdbRestore() error {
	file, err := rs.openRdbFile()
	if err != nil {
		return err
	}
	defer file.Close()

	stringObjects, err := parseRdbFile(file)
	if err != nil {
		return fmt.Errorf("error parsing RDB file: %w", err)
	}

	for _, strObj := range stringObjects {
		rs.Add(strObj.Key, string(strObj.Value))
		if strObj.Expiration != nil {
			rs.AddExpiry(strObj.Key, strObj.Expiration.UnixMilli())
		}
	}

	return nil
}

func (rs RedisStoreImpl) openRdbFile() (*os.File, error) {
	dir, ok := rs.config.Get(redisConfig.ConfigDir)
	if !ok {
		return nil, fmt.Errorf("no config value set for key %s", redisConfig.ConfigDir)
	}
	dbFilename, ok := rs.config.Get(redisConfig.ConfigDbFilename)
	if !ok {
		return nil, fmt.Errorf("no config value set for key %s", redisConfig.ConfigDbFilename)
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", dir, dbFilename))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func parseRdbFile(file *os.File) ([]parser.StringObject, error) {
	decoder := parser.NewDecoder(file)
	var stringObjects []parser.StringObject

	err := decoder.Parse(func(o parser.RedisObject) bool {
		switch o.GetType() {
		case parser.StringType:
			str := o.(*parser.StringObject)
			fmt.Printf("%s: %q \n", str.Key, str.Value)
			stringObjects = append(stringObjects, *str)
		case parser.ListType:
			list := o.(*parser.ListObject)
			println(list.Key, list.Values)
			panic("List type is not supported yet")
		case parser.HashType:
			hash := o.(*parser.HashObject)
			println(hash.Key, hash.Hash)
			panic("Hash type is not supported yet")
		case parser.ZSetType:
			zset := o.(*parser.ZSetObject)
			println(zset.Key, zset.Entries)
			panic("ZSet type is not supported yet")
		case parser.StreamType:
			stream := o.(*parser.StreamObject)
			println(stream.Entries, stream.Groups)
			panic("Stream type is not supported yet")
		}
		// return true to continue, return false to stop the iteration
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("error decoding RDB file: %w", err)
	}
	return stringObjects, nil
}
