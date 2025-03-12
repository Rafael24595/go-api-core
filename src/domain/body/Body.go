package body

type Body struct {
	Status      bool        `json:"status"`
	ContentType ContentType `json:"content_type"`
	Bytes       []byte      `json:"bytes"`
}

func NewBodyString(status bool, contentType ContentType, payload string) *Body {
	return NewBody(status, contentType, []byte(payload))
}

func NewBody(status bool, contentType ContentType, bytes []byte) *Body {
	return &Body{
		Status:      status,
		ContentType: contentType,
		Bytes:       bytes,
	}
}

func (b Body) Empty() bool {
	return b.ContentType == None
}
