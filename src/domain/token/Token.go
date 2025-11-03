package token

type Token struct {
	Id          string   `json:"id"`
	Timestamp   int64    `json:"timestamp"`
	Expire      int64    `json:"expire"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Token       string   `json:"token"`
	Scopes      []string `json:"scopes"`
	Owner       string   `json:"owner"`
}

func ToToken(hash string, token LiteToken) Token {
	return Token{
		Id:          token.Id,
		Timestamp:   token.Timestamp,
		Expire:      token.Expire,
		Name:        token.Name,
		Description: token.Description,
		Token:       hash,
		Scopes:      token.Scopes,
		Owner:       token.Owner,
	}
}

func (r Token) PersistenceId() string {
	return r.Id
}

type LiteToken struct {
	Id          string   `json:"id"`
	Timestamp   int64    `json:"timestamp"`
	Expire      int64    `json:"expire"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Scopes      []string `json:"scopes"`
	Owner       string   `json:"owner"`
}

func ToLiteToken(token Token) LiteToken {
	return LiteToken{
		Id:          token.Id,
		Timestamp:   token.Timestamp,
		Expire:      token.Expire,
		Name:        token.Name,
		Description: token.Description,
		Scopes:      token.Scopes,
		Owner:       token.Owner,
	}
}
