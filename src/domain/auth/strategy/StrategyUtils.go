package auth_strategy

import (
	"fmt"

	"github.com/Rafael24595/go-api-core/src/domain"
)

func applyHeaderAuth(key, prefix, token string, r *domain.Request) *domain.Request {
	if prefix != "" {
		prefix = fmt.Sprintf("%s ", prefix)
	}
	r.Header.Add(key, fmt.Sprintf("%s%s", prefix, token))
	return r
}
