package rdbRestorer

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/hdt3213/rdb/encoder"
	"github.com/hdt3213/rdb/parser"
)

var emptyRdbBase64 string = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="

type RedisRdbRestorer interface {
	RestoreFromRdb(filepath string) error
	SaveStoreToRdb() ([]byte, error)
}

type RdbRestorer struct {
	store *store.RedisStore
}

func NewRdbRestorer(store *store.RedisStore) (RedisRdbRestorer, error) {
	if store == nil {
		return nil, fmt.Errorf("store cannot be nil")
	}
	return &RdbRestorer{
		store: store,
	}, nil
}

func (r *RdbRestorer) RestoreFromRdb(filepath string) error {
	file, err := r.openRdbFile(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	stringObjects, err := r.parseRdbFile(file)
	if err != nil {
		return fmt.Errorf("error parsing RDB file: %w", err)
	}

	for _, strObj := range stringObjects {
		storeValue := store.NewRedisString(string(strObj.Value))
		r.store.Add(strObj.Key, storeValue)
		if strObj.Expiration != nil {
			r.store.AddExpiry(strObj.Key, strObj.Expiration.UnixMilli())
		}
	}

	return nil
}

func (r *RdbRestorer) openRdbFile(filepath string) (*os.File, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (r *RdbRestorer) parseRdbFile(file *os.File) ([]parser.StringObject, error) {
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

// Returns an empty RDB file with a header and some AUX fields
func (r *RdbRestorer) SaveStoreToRdb() ([]byte, error) {

	if r.store.Size() == 0 {
		emptyRdbBinary, err := base64.StdEncoding.DecodeString(emptyRdbBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 string: %w", err)
		}
		return emptyRdbBinary, nil
	}

	bytes, err := r.createAndWriteRdbFile()
	if err != nil {
		return nil, fmt.Errorf("failed to create and write RDB file: %w", err)
	}

	return bytes, nil
}

func (r *RdbRestorer) createAndWriteRdbFile() ([]byte, error) {
	rdbFile, err := os.Create("dumb.rdb")
	if err != nil {
		return nil, fmt.Errorf("failed to create RDB file: %w", err)
	}
	defer rdbFile.Close()
	err = rdbFile.Truncate(0) // Clear the file content
	if err != nil {
		return nil, fmt.Errorf("failed to truncate RDB file: %w", err)
	}

	enc := encoder.NewEncoder(rdbFile)
	err = enc.WriteHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to write RDB header: %w", err)
	}
	auxMap := map[string]string{
		"redis-ver":    "4.0.6",
		"redis-bits":   "64",
		"aof-preamble": "0",
	}
	for k, v := range auxMap {
		err = enc.WriteAux(k, v)
		if err != nil {
			return nil, fmt.Errorf("failed to write AUX field %s: %w", k, err)
		}
	}

	var dbIndex uint = 0
	var keyCount uint64 = 0
	var ttlCount uint64 = 0
	err = enc.WriteDBHeader(dbIndex, keyCount, ttlCount)
	if err != nil {
		return nil, fmt.Errorf("failed to write DB header: %w", err)
	}

	err = enc.WriteEnd()
	if err != nil {
		return nil, fmt.Errorf("failed to write RDB end: %w", err)
	}

	bytes, err := io.ReadAll(rdbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read RDB file: %w", err)
	}

	return bytes, nil
}
