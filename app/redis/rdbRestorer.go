package redis

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/hdt3213/rdb/parser"
)

type RedisRdbRestorer interface {
	RestoreFromRdb(filepath string) error
}

type RdbRestorer struct {
	store store.RedisStore
}

func NewRdbRestorer(store store.RedisStore) RedisRdbRestorer {
	return &RdbRestorer{
		store: store,
	}
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
		r.store.Add(strObj.Key, string(strObj.Value))
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
