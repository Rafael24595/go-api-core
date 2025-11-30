package auth

type Auths struct {
	Status bool            `json:"status"`
	Auths  map[string]Auth `json:"auths"`
}

func NewAuths(status bool) *Auths {
	return &Auths{
		Status: status,
		Auths:  make(map[string]Auth),
	}
}

func (a *Auths) PutAuth(auth Auth) *Auths {
	a.Auths[auth.Type.String()] = auth
	return a
}
