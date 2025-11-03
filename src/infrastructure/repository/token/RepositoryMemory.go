package token

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	token_domain "github.com/Rafael24595/go-api-core/src/domain/token"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
	"github.com/google/uuid"
)

type RepositoryMemory struct {
	muMemory   sync.RWMutex
	muFile     sync.RWMutex
	collection collection.IDictionary[string, token_domain.Token]
	file       repository.IFileManager[token_domain.Token]
}

func InitializeRepositoryMemory(impl collection.IDictionary[string, token_domain.Token], file repository.IFileManager[token_domain.Token]) (*RepositoryMemory, error) {
	requests, err := file.Read()
	if err != nil {
		return nil, err
	}

	return &RepositoryMemory{
		collection: impl.Merge(collection.DictionaryFromMap(requests)),
		file:       file,
	}, nil
}

func (r *RepositoryMemory) FindAll(owner string) []token_domain.LiteToken {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	tokens := r.collection.ValuesVector().
		Filter(func(t token_domain.Token) bool {
			return t.Owner == owner
		})
	return collection.VectorMap(tokens, func(t token_domain.Token) token_domain.LiteToken {
		return token_domain.ToLiteToken(t)
	}).Collect()
}

func (r *RepositoryMemory) Find(id string) (*token_domain.Token, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.Get(id)
}

func (r *RepositoryMemory) FindByName(owner, name string) (*token_domain.Token, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.FindOne(func(s string, t token_domain.Token) bool {
		return t.Owner == owner && t.Name == name
	})
}

func (r *RepositoryMemory) FindByToken(owner, token string) (*token_domain.Token, bool) {
	r.muMemory.RLock()
	defer r.muMemory.RUnlock()
	return r.collection.FindOne(func(s string, t token_domain.Token) bool {
		return t.Owner == owner && t.Token == token
	})
}

func (r *RepositoryMemory) Insert(owner string, token *token_domain.Token) *token_domain.Token {
	r.muMemory.Lock()
	return r.resolve(owner, token)
}

func (r *RepositoryMemory) resolve(owner string, token *token_domain.Token) *token_domain.Token {
	if token.Id != "" {
		return r.insert(owner, token)
	}

	key := uuid.New().String()
	if r.collection.Exists(key) {
		return r.resolve(owner, token)
	}

	token.Id = key

	return r.insert(owner, token)
}

func (r *RepositoryMemory) insert(owner string, token *token_domain.Token) *token_domain.Token {
	defer r.muMemory.Unlock()

	token.Owner = owner

	if token.Timestamp == 0 {
		token.Timestamp = time.Now().UnixMilli()
	}

	r.collection.Put(token.Id, *token)

	go r.write(r.collection)

	return token
}

func (r *RepositoryMemory) Delete(endPoint *token_domain.Token) *token_domain.Token {
	r.muMemory.Lock()
	defer r.muMemory.Unlock()

	cursor, _ := r.collection.Remove(endPoint.Id)
	go r.write(r.collection)

	return cursor
}

func (r *RepositoryMemory) write(snapshot collection.IDictionary[string, token_domain.Token]) {
	r.muFile.Lock()
	defer r.muFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v token_domain.Token) any {
		return v
	}).Values()

	err := r.file.Write(items)
	if err != nil {
		log.Error(err)
	}
}
