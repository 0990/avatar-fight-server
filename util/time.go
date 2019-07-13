package util

import "time"

func GetCurrentSec() int64 {
	return time.Now().Unix()
}

func GetCurrentMillSec() int64 {
	return time.Now().UnixNano() / 1e6
}
