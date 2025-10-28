package auth_strategy

import (
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
)

func ApplyAuth(req *action.Request) *action.Request {
	if !req.Auth.Status {
		return req
	}

	for _, a := range req.Auth.Auths {
		if !a.Status {
			continue
		}
		strategy := LoadStrategy(a.Type)
		req = strategy(a, req)
	}

	return req
}

func LoadStrategy(typ auth.Type) func(a auth.Auth, r *action.Request) *action.Request {
	switch typ {
	case auth.Basic:
		return applyBasicAuth
	case auth.Bearer:
		return applyBearerAuth
	default:
		return applyVoidAuth
	}
}
