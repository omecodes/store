package utime

import "time"

func Now() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}
