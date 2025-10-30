package utils

import (
	"fmt"
	"time"
)

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

func SplitCommand(input string) ([]string, error) {
	args := make([]string, 0)
	buffer := make([]rune, 0)

	quote := rune(0)
	escaped := false
	closed := false
	
	for _, r := range input {
		switch {
		case escaped:
			buffer = append(buffer, r)
			escaped = false
			closed = false

		case r == '\\':
			escaped = true
			closed = false

		case quote != 0:
			if r == quote {
				quote = 0
				closed = true
			} else {
				buffer = append(buffer, r)
				closed = false
			}

		case r == '"' || r == '\'':
			quote = r
			closed = false

		case r == ' ' || r == '\t' || r == '\n':
			if len(buffer) > 0 || closed {
				args = append(args, string(buffer))
				buffer = make([]rune, 0)
			}
			closed = false

		default:
			buffer = append(buffer, r)
			closed = false
		}
	}

	if len(buffer) > 0 || closed {
		args = append(args, string(buffer))
	}

	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote")
	}

	return args, nil
}
