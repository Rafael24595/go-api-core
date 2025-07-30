package header

type Headers struct {
	Headers map[string][]Header `json:"headers"`
}

func NewHeaders() *Headers {
	return &Headers{
		Headers: make(map[string][]Header),
	}
}

func (h *Headers) Add(key, value string) *Headers {
	return h.AddHeader(key, Header{
		Order: int64(len(h.Headers)),
		Status: true,
		Value: value,
	})
}

func (h *Headers) AddHeader(key string, header Header) *Headers {
	if _, ok := h.Headers[key]; !ok {
		h.Headers[key] = make([]Header, 0)
	}

	h.Headers[key] = append(h.Headers[key], header)

	return h
}

func (q *Headers) SizeOf(key string) int {
	if headers, ok := q.Headers[key]; ok {
		return len(headers)
	}
	return 0
}
