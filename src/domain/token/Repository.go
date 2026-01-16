package token

type Repository interface {
	FindAll(owner string) []LiteToken
	Find(id string) (*Token, bool)
	FindGlobal(token string) (*Token, bool)
	FindByName(owner, name string) (*Token, bool)
	FindByToken(owner, token string) (*Token, bool)
	Insert(owner string, token *Token) *Token
	Delete(token *Token) *Token
}
