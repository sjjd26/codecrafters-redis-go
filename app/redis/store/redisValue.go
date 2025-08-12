package store

import "fmt"

type RedisType int

const (
	RedisString RedisType = iota
	RedisList
)

type RedisValue struct {
	Type  RedisType
	Value any
}

func NewRedisString(value string) RedisValue {
	return RedisValue{
		Type:  RedisString,
		Value: value,
	}
}

func NewRedisList(value []string) RedisValue {
	return RedisValue{
		Type:  RedisList,
		Value: value,
	}
}

func ExtractString(value RedisValue) (string, error) {
	if value.Type != RedisString {
		return "", fmt.Errorf("value is not a string")
	}
	str, ok := value.Value.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string")
	}
	return str, nil
}

func ExtractList(value RedisValue) ([]string, error) {
	if value.Type != RedisList {
		return nil, fmt.Errorf("value is not a list")
	}
	list, ok := value.Value.([]string)
	if !ok {
		return nil, fmt.Errorf("value is not a list")
	}
	return list, nil
}
