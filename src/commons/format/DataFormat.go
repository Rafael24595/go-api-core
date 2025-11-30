package format

import "strings"

type DataFormat string

const (
	CSVT DataFormat = "csvt"
)

var allDataFormat = []DataFormat{
	CSVT,
}

func DataFormatFromString(s string) (DataFormat, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, t := range allDataFormat {
		if string(t) == s {
			return t, true
		}
	}
	return "", false
}

func DataFormatFromExtension(s string) (DataFormat, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, ".")
	for _, t := range allDataFormat {
		if t.Extension() == s {
			return t, true
		}
	}
	return "", false
}

func (f DataFormat) Extension() string {
	return "csvt"
}
