package body

type Body struct {
	ContentType ContentType `json:"content_type"`
	Bytes       []byte      `json:"bytes"`
}