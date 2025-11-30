package cookie

type CookiesClient struct {
	Cookies map[string]CookieClient `json:"cookies"`
}

func NewCookiesClient() *CookiesClient {
	return &CookiesClient{
		Cookies: make(map[string]CookieClient),
	}
}

func (h *CookiesClient) Find(key string) (*CookieClient, bool) {
	cookie, ok := h.Cookies[key]
	return &cookie, ok
}


func (h *CookiesClient) Put(key, value string) *CookiesClient {
	return h.PutStatus(key, value, true)
}

func (h *CookiesClient) PutStatus(key, value string, status bool) *CookiesClient {
	h.Cookies[key] = CookieClient{
		Order: int64(len(h.Cookies)),
		Status: status,
		Value: value,
	}
	return h
}

type CookiesServer struct {
	Cookies map[string]CookieServer `json:"cookies"`
}

func NewCookiesServer() *CookiesServer {
	return &CookiesServer{
		Cookies: make(map[string]CookieServer),
	}
}
