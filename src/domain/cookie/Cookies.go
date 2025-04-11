package cookie

type CookiesClient struct {
	Cookies map[string]CookieClient `json:"cookies"`
}

func NewCookiesClient() *CookiesClient {
	return &CookiesClient{
		Cookies: make(map[string]CookieClient),
	}
}

type CookiesServer struct {
	Cookies map[string]CookieServer `json:"cookies"`
}

func NewCookiesServer() *CookiesServer {
	return &CookiesServer{
		Cookies: make(map[string]CookieServer),
	}
}