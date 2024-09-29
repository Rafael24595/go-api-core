package auth

type Auth struct {
	Active     bool                 `json:"active"`
	Type       Type                 `json:"type"`
	Parameters map[string]Parameter `json:"parameters"`
}

func NewAuthEmpty(active bool, typ Type) *Auth {
	return NewAuth(active, typ, make(map[string]Parameter))
}

func NewAuth(active bool, typ Type, parameters map[string]Parameter) *Auth {
	return &Auth{
		Active: active,
		Type: typ,
		Parameters: parameters,
	}
}

func (a *Auth) PutParam(key, value string) *Auth {
	a.Parameters[key] = *NewParameter(key, value)
	return a
}
