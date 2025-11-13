package swr

import "regexp"

func findValue(cursor string) (string, bool) {
	re := regexp.MustCompile(REGEX_RAW_VALUE)
	matches := re.FindStringSubmatch(cursor)
	if len(matches) < 2 {
		return "", false
	}
	return matches[1], true
}

func findPosition(cursor string) (string, bool) {
	re := regexp.MustCompile(REGEX_VECTOR_INDEX)
	matches := re.FindStringSubmatch(cursor)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}
