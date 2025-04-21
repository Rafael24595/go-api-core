package cookie

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type CookieClient struct {
	Order  int64  `json:"order"`
	Status bool   `json:"status"`
	Value  string `json:"value"`
}

func NewCookieClient(order int64, status bool, value string) CookieClient {
	return CookieClient{
		Order:  order,
		Status: status,
		Value:  value,
	}
}

type CookieServer struct {
	Status     bool     `json:"status"`
	Code       string   `json:"code"`
	Value      string   `json:"value"`
	Domain     string   `json:"domain"`
	Path       string   `json:"path"`
	Expiration string   `json:"expiration"`
	MaxAge     int      `json:"maxage"`
	Secure     bool     `json:"secure"`
	HttpOnly   bool     `json:"httponly"`
	SameSite   SameSite `json:"samesite"`
}

func CookieServerFromString(cookieString string) (*CookieServer, error) {
	parts := strings.Split(cookieString, ";")

	codeValue := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(codeValue) != 2 {
		return nil, errors.New("invalid cookie format")
	}

	code := strings.TrimSpace(codeValue[0])
	value := strings.TrimSpace(codeValue[1])

	cookie := &CookieServer{
		Status:   true,
		Code:     code,
		Value:    value,
		Secure:   false,
		HttpOnly: false,
	}

	for _, part := range parts[1:] {
		keyValue := strings.SplitN(strings.TrimSpace(part), "=", 2)
		key := strings.ToLower(strings.TrimSpace(keyValue[0]))

		var value string
		if len(keyValue) > 1 {
			value = strings.TrimSpace(keyValue[1])
		}

		switch key {
		case "secure":
			cookie.Secure = true
		case "httponly":
			cookie.HttpOnly = true
		case "expires":
			cookie.Expiration = value
		case "domain":
			cookie.Domain = value
		case "path":
			cookie.Path = value
		case "max-age":
			if value != "" {
				maxAge, err := strconv.Atoi(value)
				if err != nil {
					return nil, errors.New("invalid Max-Age value")
				}
				cookie.MaxAge = maxAge
			}
		case "samesite":
			if value != "" {
				sameSite, err := SameSiteFromString(value)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("unknown SameSite value: '%s'", value))
				}
				cookie.SameSite = *sameSite
			}
		default:
			return nil, errors.New(fmt.Sprintf("unknown field code: '%s'", key))
		}
	}

	return cookie, nil
}

func (c *CookieServer) String() string {
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
