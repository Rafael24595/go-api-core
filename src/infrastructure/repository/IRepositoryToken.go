package repository

import "github.com/Rafael24595/go-api-core/src/domain/token"

type IRepositoryToken interface {
	FindAll(owner string) []token.LiteToken
	Find(id string) (*token.Token, bool)
	FindGlobal(token string) (*token.Token, bool)
	FindByName(owner, name string) (*token.Token, bool)
	FindByToken(owner, token string) (*token.Token, bool)
	Insert(owner string, token *token.Token) *token.Token
	Delete(token *token.Token) *token.Token
}
