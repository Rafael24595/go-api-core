package auth

import (
	"fmt"
	"net/http"
)

func applyHeaderAuth(key, prefix, token string, r *http.Request) *http.Request {
	q := r.URL.Query()
	if prefix != "" {
		prefix = fmt.Sprintf("%s ", prefix)
	}
	q.Add(key, fmt.Sprintf("%s%s", prefix, token))
	r.URL.RawQuery = q.Encode()
	return r	
}