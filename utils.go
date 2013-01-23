package main

import (
	"time"
)

func Timestamp() int64{
	return time.Now().Unix()
}
