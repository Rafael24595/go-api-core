package header

type Headers struct {
	Headers map[string]Header `json:"headers"`
}

func NewHeaders() *Headers {
	return &Headers{
		Headers: make(map[string]Header),
	}
}

func (h *Headers) Add(header Header) *Headers {
	param, ok := h.Headers[header.Key]
	if !ok {
		h.Headers[header.Key] = header
		return h
	}

	param.Header = append(param.Header, header.Header...)
	
	if header.Active && !param.Active {
		param.Active = header.Active
	}

	return h
}
