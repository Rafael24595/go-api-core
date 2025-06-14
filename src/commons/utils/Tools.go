package utils

import "time"

func FormatMilliseconds(timestamp int64) string {
	if timestamp == 0 {
		return "N/A"
	}
	seconds := timestamp / 1000
	time := time.Unix(seconds, 0)
	return time.Format("2006-01-02 15:04:05")
}

func FormatMillisecondsCompact(timestamp int64) string {
	if timestamp == 0 {
		return "N/A"
	}
	seconds := timestamp / 1000
	time := time.Unix(seconds, 0)
	return time.Format("20060102_150405")
}
