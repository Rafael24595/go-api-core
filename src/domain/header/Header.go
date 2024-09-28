package header

type Header struct {
	Active bool     `json:"active"`
	Key    string   `json:"key"`
	Header []string `json:"header"`
}

func NewHeader(active bool, key string, header ...string) Header {
	return Header{
		Active: active,
		Key: key,
		Header: header,
	}
}
