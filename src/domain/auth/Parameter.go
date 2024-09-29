package auth

type Parameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewParameter(key, value string) *Parameter {
	return &Parameter{
		Key: key,
		Value: value,
	}
}