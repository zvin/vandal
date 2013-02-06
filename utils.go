package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"math"
	"reflect"
	"sync"
	"time"
)

type LockableWebsiteSlice struct {
	Mutex sync.RWMutex
	Sites []Website
}

func Base64Encode(data string) string {
	bb := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, bb)
	encoder.Write([]byte(data))
	encoder.Close()
	return bb.String()
}

func MaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func Distance(x1, y1, x2, y2 int) float64 {
	return math.Sqrt(float64((x1 - x2) ^ 2 + (y1 - y2) ^ 2))
}

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
