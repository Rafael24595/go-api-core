package cookie

type Cookies struct {
	Cookies map[string]Cookie `json:"cookies"`
}

func NewCookies() *Cookies {
	return &Cookies{
		Cookies: make(map[string]Cookie),
	}
}