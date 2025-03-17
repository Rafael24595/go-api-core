package header

type Headers struct {
	Headers map[string][]Header `json:"headers"`
}

func NewHeaders() *Headers {
	return &Headers{
		Headers: make(map[string][]Header),
	}
}

func (h *Headers) Add(key string, header Header) *Headers {
	if _, ok := h.Headers[key]; !ok {
		h.Headers[key] = make([]Header, 0)
	}

	h.Headers[key] = append(h.Headers[key], header)

	return h
}
