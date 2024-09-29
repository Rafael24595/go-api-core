package auth

type Auths struct {
	Auths map[string]Auth `json:"auths"`
}

func NewAuths() *Auths {
	return &Auths{
		Auths: make(map[string]Auth),
	}
}

func (a *Auths) PutAuth(auth Auth) *Auths {
	a.Auths[auth.Type.String()] = auth
	return a
}