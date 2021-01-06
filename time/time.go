package time

import "time"

func TimestampNowMs() int64 {
	return time.Now().UTC().UnixNano() / 1000000
}
