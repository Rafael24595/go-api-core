package header

type Header struct {
	Status bool   `json:"status"`
	Value  string `json:"value"`
}

func NewHeader(status bool, value string) Header {
	return Header{
		Status: status,
		Value:  value,
	}
}
