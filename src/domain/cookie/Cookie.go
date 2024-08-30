package cookie

import (
	"fmt"
	"go-api-core/src/commons"
	"strconv"
	"strings"
)

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

func CookieFromString(cookieString string) (*Cookie, error) {
	parts := strings.Split(cookieString, ";")

	codeValue := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(codeValue) != 2 {
		return nil, commons.ApiErrorFrom(422, "Invalid cookie format")
	}

	code := strings.TrimSpace(codeValue[0])
	value := strings.TrimSpace(codeValue[1])

	cookie := &Cookie{
		Code: code,
		Value: value,
		Secure: false,
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
					return nil, commons.ApiErrorFromCause(422, "Invalid Max-Age value", err)
				}
				cookie.MaxAge = maxAge
			}
		case "samesite":
			if value != "" {
				sameSite, err := SameSiteFromString(value)
				if err != nil {
					return nil,  commons.ApiErrorFromCause(422, fmt.Sprintf("Unknown SameSite value: '%s'", value), err)
				}
				cookie.SameSite = *sameSite
			}
		default:
			return nil, commons.ApiErrorFrom(422, fmt.Sprintf("Unknown field code: '%s'", key))
		}
	}

	return cookie, nil
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