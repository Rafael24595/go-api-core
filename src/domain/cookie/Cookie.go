package cookie

import "fmt"

type Cookie struct {
	Code       string
	Value      string
	Domain     string
	Path       string
	Expiration string
	MaxAge     int
	Secure     bool
	HttpOnly   bool
	SameSite   SameSite
}

func (c *Cookie) String() string {
	cookieString := c.Value

	if c.Domain != "" {
		cookieString += fmt.Sprintf("; Domain=%s", c.Domain)
	}

	if c.Path != "" {
		cookieString += fmt.Sprintf("; Path=%s", c.Path)
	}

	if c.Expiration != "" {
		cookieString += fmt.Sprintf("; Expires=%s", c.Expiration)
	}

	if c.MaxAge != 0 {
		cookieString += fmt.Sprintf("; Max-Age=%d", c.MaxAge)
	}

	if c.Secure {
		cookieString += "; Secure"
	}

	if c.HttpOnly {
		cookieString += "; HttpOnly"
	}

	if c.SameSite != None {
		cookieString += fmt.Sprintf("; SameSite=%s", c.SameSite)
	}

	return cookieString
}