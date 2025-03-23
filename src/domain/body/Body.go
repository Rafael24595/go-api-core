package body

type Body struct {
	Status      bool        `json:"status"`
	ContentType ContentType `json:"content_type"`
	Payload     []byte      `json:"payload"`
}

func NewBodyString(status bool, contentType ContentType, payload string) *Body {
	return NewBody(status, contentType, []byte(payload))
}

func NewBody(status bool, contentType ContentType, bytes []byte) *Body {
	return &Body{
		Status:      status,
		ContentType: contentType,
		Payload:     bytes,
	}
}

func (b Body) Empty() bool {
	return b.ContentType == None
}
