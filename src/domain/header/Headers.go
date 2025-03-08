package header

type Headers struct {
	Headers map[string][]Header `json:"headers"`
}

func NewHeaders() *Headers {
	return &Headers{
		Headers: make(map[string][]Header),
	}
}

func (h *Headers) Add(header Header) *Headers {
	if _, ok := h.Headers[header.Key]; !ok {
		h.Headers[header.Key] = make([]Header, 0)
	}

	h.Headers[header.Key] = append(h.Headers[header.Key], header)

	return h
}
