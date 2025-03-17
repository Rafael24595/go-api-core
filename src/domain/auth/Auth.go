package auth

type Auth struct {
	Status     bool                 `json:"status"`
	Type       Type                 `json:"type"`
	Parameters map[string]Parameter `json:"parameters"`
}

func NewAuthEmpty(status bool, typ Type) *Auth {
	return NewAuth(status, typ, make(map[string]Parameter))
}

func NewAuth(status bool, typ Type, parameters map[string]Parameter) *Auth {
	return &Auth{
		Status: status,
		Type: typ,
		Parameters: parameters,
	}
}

func (a *Auth) PutParam(key, value string) *Auth {
	a.Parameters[key] = *NewParameter(key, value)
	return a
}
