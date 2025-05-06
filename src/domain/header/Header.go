package header

type Header struct {
	Order  int64  `json:"order"`
	Status bool   `json:"status"`
	Value  string `json:"value"`
}

func NewHeader(order int64, status bool, value string) Header {
	return Header{
		Order:  order,
		Status: status,
		Value:  value,
	}
}
