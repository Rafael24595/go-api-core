package cookie

type CookiesClient struct {
	Cookies map[string]CookieClient `json:"cookies"`
}

func NewCookiesClient() *CookiesClient {
	return &CookiesClient{
		Cookies: make(map[string]CookieClient),
	}
}

func (h *CookiesClient) Put(key, value string) *CookiesClient {
	h.Cookies[key] = CookieClient{
		Order: int64(len(h.Cookies)),
		Status: true,
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
