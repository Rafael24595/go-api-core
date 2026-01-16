package manager

import (
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain/token"
)

type ManagerToken struct {
	token token.Repository
}

func NewManagerToken(token token.Repository) *ManagerToken {
	return &ManagerToken{
		token: token,
	}
}

func (m *ManagerToken) FindAll(owner string) []token.LiteToken {
	return m.token.FindAll(owner)
}

func (m *ManagerToken) FindGlobal(tkn string) (*token.Token, bool) {
	conf := configuration.Instance()
	hash := token.HashToken(string(conf.Secret()), tkn)
	return m.token.FindGlobal(hash)
}

func (m *ManagerToken) FindByName(owner, name string) (*token.Token, bool) {
	return m.token.FindByName(owner, name)
}

func (m *ManagerToken) FindByToken(owner, tkn string) (*token.Token, bool) {
	conf := configuration.Instance()
	hash := token.HashToken(string(conf.Secret()), tkn)
	return m.token.FindByToken(owner, hash)
}

func (m *ManagerToken) Insert(owner string, tkn *token.LiteToken) (string, *token.LiteToken) {
	conf := configuration.Instance()

	if tkn.Owner == "" {
		tkn.Owner = owner
	}

	if tkn.Owner != owner {
		return "", nil
	}

	retries := 5
	for range retries {
		raw := token.GenerateRawToken()
		hash := token.HashToken(string(conf.Secret()), raw)

		if _, exists := m.token.FindGlobal(hash); exists {
			continue
		}

		tkn := token.ToToken(hash, *tkn)
		result := m.token.Insert(owner, &tkn)
		if result == nil {
			return "", nil
		}

		lite := token.ToLiteToken(*result)
		return raw, &lite
	}

	log.Warningf("failed to generate a unique token for user %s after %d attempts", owner, retries)

	return "", nil
}

func (m *ManagerToken) DeleteById(owner string, token string) (*token.Token, bool) {
	tkn, exists := m.token.Find(token)
	if !exists {
		return nil, false
	}

	return m.token.Delete(tkn), true
}

func (m *ManagerToken) Delete(owner string, token *token.Token) *token.Token {
	if token.Owner != owner {
		return nil
	}

	return m.token.Delete(token)
}
