package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/session"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-collections/collection"
	"golang.org/x/crypto/bcrypt"
)

var manager *ManagerSession

type ManagerSession struct {
	mut            sync.RWMutex
	mutFile        sync.RWMutex
	file           IFileManager[dto.DtoSession]
	contextManager *ManagerContext
	sessions       collection.IDictionary[string, session.Session]
}

func InitializeManagerSession(file IFileManager[dto.DtoSession], contextManager *ManagerContext) (*ManagerSession, error) {
	if manager != nil {
		return nil, errors.New("already instanced")
	}

	steps, err := file.Read()
	if err != nil {
		return nil, err
	}

	sessions := collection.DictionaryMap(collection.DictionaryFromMap(steps), func(k string, d dto.DtoSession) session.Session {
		return *dto.ToSession(d)
	})

	instance := &ManagerSession{
		file:           file,
		contextManager: contextManager,
		sessions:       sessions,
	}

	conf := configuration.Instance()
	err = instance.defineDefaultUser(conf.Admin(), string(conf.Secret()))
	if err != nil {
		panic(err)
	}

	err = instance.defineDefaultUser("anonymous", "")
	if err != nil {
		panic(err)
	}

	manager = instance

	return manager, nil
}

func InstanceManagerSession() *ManagerSession {
	if manager == nil {
		panic("Not initialized")
	}
	return manager
}

func (s *ManagerSession) defineDefaultUser(username, secret string) error {
	if _, exists := s.sessions.Get(username); !exists {
		ctx := s.contextManager.Insert(username, context.NewContext(username))
		_, err := s.Insert(username, string(secret), ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ManagerSession) Find(user string) (*session.Session, bool) {
	return s.sessions.Get(user)
}

func (s *ManagerSession) Authorize(user, password string) (*session.Session, error) {
	session, exists := s.sessions.Get(user)
	if !exists {
		return nil, errors.New("session not found")
	}

	if !ValideSecret([]byte(password), session.Secret) {
		return nil, errors.New("session not found")
	}

	return session, nil
}

func (s *ManagerSession) Insert(user, password string, ctx *context.Context) (*session.Session, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	_, exists := s.sessions.Get(user)
	if exists {
		return nil, errors.New("user already exists")
	}

	secret, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	session := session.Session{
		Username:  user,
		Secret:    secret,
		Timestamp: time.Now().UnixMilli(),
		Context:   ctx.Id,
	}

	s.sessions.Put(user, session)

	go s.write(s.sessions)

	return &session, nil
}

func (s *ManagerSession) write(snapshot collection.IDictionary[string, session.Session]) {
	s.mutFile.Lock()
	defer s.mutFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v session.Session) any {
		return *dto.FromSession(v)
	}).Values()

	err := s.file.Write(items)
	if err != nil {
		println(err.Error())
	}
}

func HashPassword(password string) ([]byte, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return hashedBytes, err
}

func ValideSecret(password, hashed []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashed, password)
	return err == nil
}
