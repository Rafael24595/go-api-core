package swr

const (
	// REGEX_VECTOR_INDEX matches array/vector index expressions like [0], [1], etc.
	REGEX_VECTOR_INDEX = `^\[(.*)\]$`
	// REGEX_RAW_VALUE matches raw literal values wrapped in angle brackets like <value>.
	REGEX_RAW_VALUE = `^<(.+)>$`
)
