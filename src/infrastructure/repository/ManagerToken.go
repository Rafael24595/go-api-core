package repository

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
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

func (m *ManagerToken) FindByName(owner, name string) (*token_domain.Token, bool) {
	return m.token.FindByName(owner, name)
}

func (m *ManagerToken) FindByToken(owner, token string) (*token_domain.Token, bool) {
	conf := configuration.Instance()
	hash := m.hashToken(string(conf.Secret()), token)
	return m.token.FindByToken(owner, hash)
}

func (m *ManagerToken) Insert(owner string, token *token_domain.LiteToken) (string, *token_domain.LiteToken) {
	conf := configuration.Instance()

	if token.Owner != owner {
		return "", nil
	}

	raw := m.generateRawToken()
	hash := m.hashToken(string(conf.Secret()), raw)

	tkn := token_domain.ToToken(hash, *token)
	result := m.token.Insert(owner, &tkn)
	if result == nil {
		return "", nil
	}

	lite := token_domain.ToLiteToken(*result)
	return raw, &lite
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

func (m *ManagerToken) generateRawToken() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func (m *ManagerToken) hashToken(secret, token string) string {
	tkn := fmt.Sprintf("%s:%s", secret, token)
	sum := sha256.Sum256([]byte(tkn))
	return hex.EncodeToString(sum[:])
}
