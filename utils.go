package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type LockableWebsiteSlice struct {
	Mutex sync.RWMutex
	Sites []*Website
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
	return math.Sqrt(math.Pow(float64(x1-x2), 2) + math.Pow(float64(y1-y2), 2))
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

func ToString(n interface{}) (result string, err error) {
	switch n.(type) {
	case []uint8:
		result = string(n.([]uint8))
	default:
		Log.Printf("ToString, not a string: %#v ", n)
		err = errors.New("Not a string")
	}
	return result, err
}

func Pluralize(s string, nb int) string {
	if nb > 1 {
		return s + "s"
	} else {
		return s
	}
}

func TruncateString(s string, nb int) string {
	length := utf8.RuneCountInString(s)
	if length <= nb {
		return s
	} else {
		letters := strings.Split(s, "")
		letters = append(letters[:nb-1], "…")
		return strings.Join(letters, "")
	}
}

func TryQueryUnescape(s string) string {
	label, err := url.QueryUnescape(s)
	if err != nil {
		Log.Printf("Unescaping url %s failed\n", s)
		label = s
	}
	return label
}

func ReadIntFromFile(filename string) uint32 {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return 1
	}
	value, err := strconv.Atoi(strings.Trim(string(content), "\n"))
	if err != nil {
		return 1
	}
	return uint32(value)
}

func WriteIntToFile(value uint32, filename string) {
	ioutil.WriteFile(filename, []byte(fmt.Sprintf("%d", value)), 0600)
}
