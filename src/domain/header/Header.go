package header

type Header struct {
	Active bool   `json:"active"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func NewHeader(active bool, key string, value string) Header {
	return Header{
		Active: active,
		Key:    key,
		Value:  value,
	}
}
