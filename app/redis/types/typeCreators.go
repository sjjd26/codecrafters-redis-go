package types

import "fmt"

const BulkStringNull = "$-1\r\n"
const OkString = "+OK\r\n"
const EmptyArray = "*0\r\n"

func CreateBulkString(str string) string {
	return fmt.Sprintf("$%v\r\n%s\r\n", len(str), str)
}

func CreateKeyValueArray(key, value string) string {
	return fmt.Sprintf("*2\r\n%s%s", CreateBulkString(key), CreateBulkString(value))
}

func CreateKeyValueNullArray(key string) string {
	return fmt.Sprintf("*2\r\n%s%s", CreateBulkString(key), BulkStringNull)
}

func CreateEmptyArray() string {
	return EmptyArray
}

func CreateBulkStringArray(values []string) string {
	result := fmt.Sprintf("*%d\r\n", len(values))
	for _, value := range values {
		result += CreateBulkString(value)
	}
	return result
}

func CreateInt(value int) string {
	return fmt.Sprintf(":%d\r\n", value)
}

func CreateSimpleString(value string) string {
	return fmt.Sprintf("+%s\r\n", value)
}
