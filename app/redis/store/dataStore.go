package store

import (
	"fmt"
	"time"
)

type ExpiryValue struct {
	value     string
	timeAdded time.Time
	expiry    int // milliseconds
}

var dataStore = make(map[string]ExpiryValue)

func Add(key string, value string, expiry int) {
	now := time.Now()
	expiryVal := ExpiryValue{value: value, expiry: expiry, timeAdded: now}
	dataStore[key] = expiryVal
}

func Get(key string) (string, bool) {
	expValue, ok := dataStore[key]
	if !ok {
		return "", false
	}

	now := time.Now()
	diff := now.Sub(expValue.timeAdded)
	fmt.Printf("got value %v with expiry %v, diff: %v \n", expValue.value, expValue.expiry, diff.Milliseconds())
	if int(diff.Milliseconds()) > expValue.expiry {
		delete(dataStore, key)
		return "", false
	}

	return expValue.value, true
}
