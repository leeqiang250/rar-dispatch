package time

import "time"

func TimestampNowMs() int64 {
	return time.Now().UTC().UnixNano() / 1000000
}

func Sleep(d time.Duration) {
	time.Sleep(d)
}
