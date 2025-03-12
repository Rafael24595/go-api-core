package header

type Header struct {
	Status bool   `json:"status"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func NewHeader(status bool, key string, value string) Header {
	return Header{
		Status: status,
		Key:    key,
		Value:  value,
	}
}
