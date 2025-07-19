package redis

import (
	"fmt"
)

type RespType int

var ErrUnknownTypeByte = fmt.Errorf("uknown type byte")

const (
	RespUnknown = iota
	RespSimpleStr
	RespSimpleErr
	RespInt
	RespBulkStr
	RespArray
	RespNull
	RespBool
	RespDouble
	RespBigNum
	RespBulkErr
	RespVerbatimStr
	RespMap
	RespAttribute
	RespSet
	RespPush
)

var RespTypeByteMap = map[byte]RespType{
	'+': RespSimpleStr,
	'-': RespSimpleErr,
	':': RespInt,
	'$': RespBulkStr,
	'*': RespArray,
	'_': RespNull,
	'#': RespBool,
	',': RespDouble,
	'(': RespBigNum,
	'!': RespBulkErr,
	'=': RespVerbatimStr,
	'%': RespMap,
	'|': RespAttribute,
	'~': RespSet,
	'>': RespPush,
}

func GetRespType(typeByte byte) (RespType, error) {
	if respType, ok := RespTypeByteMap[typeByte]; ok {
		return respType, nil
	}
	return RespUnknown, fmt.Errorf("%w: %q", ErrUnknownTypeByte, string(typeByte))
}

func CheckTypeByte(typeByte byte, expectedType RespType) error {
	// fmt.Printf("byte type check, %q, %v \n", typeByte, expectedType)
	respType, ok := RespTypeByteMap[typeByte]
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownTypeByte, typeByte)
	}
	if respType != expectedType {
		return fmt.Errorf("type byte (%v) does not match expected (%v)", respType, expectedType)
	}
	return nil
}
