package main

import (
	"errors"
	"reflect"
	"time"
)

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Timestamp() int64 {
	return time.Now().Unix()
}

func ToInt(n interface{}) (result int, err error) {
	switch n.(type) {
	case int, int8, int16, int32, int64:
		result = int(reflect.ValueOf(n).Int())
	case uint, uint8, uint16, uint32, uint64:
		result = int(reflect.ValueOf(n).Uint())
	default:
		Log.Printf("ToInt, not an int: %#v ", n)
		err = errors.New("Not an int")
	}
	return result, err
}
