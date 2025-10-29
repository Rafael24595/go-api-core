package utils

import "time"

const ESCAPE_RUNE = '\\'

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

func SplitByRune(s string, sep rune) []string {
	var fragments []string
	var buffer []rune

	escaped := 0
	for _, r := range s {
		if r == sep && escaped%2 == 0 {
			fragments = append(fragments, string(buffer))
			buffer = make([]rune, 0)
			continue
		}

		if r == ESCAPE_RUNE {
			escaped += 1
		} else {
			escaped = 0
		}

		if r != ESCAPE_RUNE || escaped%2 == 0 {
			buffer = append(buffer, r)
		}
	}

	return append(fragments, string(buffer))
}