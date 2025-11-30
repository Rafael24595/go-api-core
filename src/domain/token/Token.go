package token

import "time"

type ScopeCode string

const (
	ScopeAPIToken ScopeCode = "apitoken"
	ScopeMockAPI  ScopeCode = "mockapi"
)

type ScopeData Scope

type Scope struct {
	Code  ScopeCode `json:"code"`
	Title string    `json:"title"`
	Value string    `json:"value"`
}

var Scopes = map[ScopeCode]Scope{
	ScopeAPIToken: {
		Code:  ScopeAPIToken,
		Title: "Allows requests to go-api API",
		Value: "apitoken",
	},
	ScopeMockAPI: {
		Code:  ScopeMockAPI,
		Title: "Allows requests to mock API",
		Value: "mockapi",
	},
}

func ListScopes() []Scope {
	scopes := make([]Scope, 0, len(Scopes))
	for _, s := range Scopes {
		scopes = append(scopes, s)
	}
	return scopes
}

type Token struct {
	Id          string      `json:"id"`
	Timestamp   int64       `json:"timestamp"`
	Expire      int64       `json:"expire"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Token       string      `json:"token"`
	Scopes      []ScopeCode `json:"scopes"`
	Owner       string      `json:"owner"`
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

func (r Token) IsExipred() bool {
	return r.Expire < time.Now().UnixMilli()
}

func (r Token) PersistenceId() string {
	return r.Id
}

type LiteToken struct {
	Id          string      `json:"id"`
	Timestamp   int64       `json:"timestamp"`
	Expire      int64       `json:"expire"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Scopes      []ScopeCode `json:"scopes"`
	Owner       string      `json:"owner"`
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
