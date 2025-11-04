package repository

import (
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	token_domain "github.com/Rafael24595/go-api-core/src/domain/token"
)

type ManagerToken struct {
	token IRepositoryToken
}

func NewManagerToken(token IRepositoryToken) *ManagerToken {
	return &ManagerToken{
		token: token,
	}
}

func (m *ManagerToken) FindAll(owner string) []token_domain.LiteToken {
	return m.token.FindAll(owner)
}

func (m *ManagerToken) FindGlobal(token string) (*token_domain.Token, bool) {
	conf := configuration.Instance()
	hash := token_domain.HashToken(string(conf.Secret()), token)
	return m.token.FindGlobal(hash)
}

func (m *ManagerToken) FindByName(owner, name string) (*token_domain.Token, bool) {
	return m.token.FindByName(owner, name)
}

func (m *ManagerToken) FindByToken(owner, token string) (*token_domain.Token, bool) {
	conf := configuration.Instance()
	hash := token_domain.HashToken(string(conf.Secret()), token)
	return m.token.FindByToken(owner, hash)
}

func (m *ManagerToken) Insert(owner string, token *token_domain.LiteToken) (string, *token_domain.LiteToken) {
	conf := configuration.Instance()

	if token.Owner == "" {
		token.Owner = owner
	}

	if token.Owner != owner {
		return "", nil
	}

	retries := 5
	for range retries {
		raw := token_domain.GenerateRawToken()
		hash := token_domain.HashToken(string(conf.Secret()), raw)
	
		if _, exists := m.token.FindGlobal(hash); exists {
			continue
		}
	
		tkn := token_domain.ToToken(hash, *token)
		result := m.token.Insert(owner, &tkn)
		if result == nil {
			return "", nil
		}
	
		lite := token_domain.ToLiteToken(*result)
		return raw, &lite
	}

	log.Warningf("failed to generate a unique token for user %s after %d attempts", owner, retries)

	return "", nil
}

func (m *ManagerToken) DeleteById(owner string, token string) (*token_domain.Token, bool) {
	tkn, exists := m.token.Find(token)
	if !exists {
		return nil, false
	}

	return m.token.Delete(tkn), true
}

func (m *ManagerToken) Delete(owner string, token *token_domain.Token) *token_domain.Token {
	if token.Owner != owner {
		return nil
	}

	return m.token.Delete(token)
}
