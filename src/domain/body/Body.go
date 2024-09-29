package body

type Body struct {
	ContentType ContentType `json:"content_type"`
	Bytes       []byte      `json:"bytes"`
}

func NewBodyString(ContentType ContentType, payload string) *Body {
	return NewBody(ContentType, []byte(payload))
}

func NewBody(ContentType ContentType, Bytes []byte ) *Body {
	return &Body{
		ContentType: ContentType,
		Bytes: Bytes,
	}
}

func (b Body) Empty() bool {
	return b.ContentType == None
}