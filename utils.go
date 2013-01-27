package main

import (
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
