package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/session"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-collections/collection"
	"golang.org/x/crypto/bcrypt"
)

var manager *ManagerSession

type ManagerSession struct {
	mut               sync.RWMutex
	mutFile           sync.RWMutex
	file              IFileManager[dto.DtoSession]
	managerCollection *ManagerCollection
	managerGroup      *ManagerGroup
	sessions          collection.IDictionary[string, session.Session]
}

func InitializeManagerSession(file IFileManager[dto.DtoSession], managerCollection *ManagerCollection, managerGroup *ManagerGroup) (*ManagerSession, error) {
	if manager != nil {
		return nil, errors.New("already instanced")
	}

	steps, err := file.Read()
	if err != nil {
		return nil, err
	}

	sessions := collection.DictionarySyncMap(collection.DictionarySyncFromMap(steps), func(k string, d dto.DtoSession) session.Session {
		return *dto.ToSession(d)
	})

	instance := &ManagerSession{
		file:              file,
		managerCollection: managerCollection,
		managerGroup:      managerGroup,
		sessions:          sessions,
	}

	manager = defineDefaultSessions(instance)

	return manager, nil
}

func defineDefaultSessions(instance *ManagerSession) *ManagerSession {
	conf := configuration.Instance()

	err := instance.defineDefaultUser(conf.Admin(), string(conf.Secret()), true, true, 0)
	if err != nil {
		log.Panic(err)
	}

	err = instance.defineDefaultUser("anonymous", "", true, false, 0)
	if err != nil {
		log.Panic(err)
	}

	return instance
}

func InstanceManagerSession() *ManagerSession {
	if manager == nil {
		log.Panics("The session manager is not initialized yet")
	}
	return manager
}

func (s *ManagerSession) defineDefaultUser(username, secret string, isProtected, isAdmin bool, count int) error {
	if _, exists := s.sessions.Get(username); !exists {
		collection, history, group := s.makeDependencies(username)
		_, err := s.insert(username, string(secret), collection, history, group, isProtected, isAdmin, count)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ManagerSession) Find(user string) (*session.Session, bool) {
	return s.sessions.Get(user)
}

func (s *ManagerSession) FindUserCollection(user string) (*domain.Collection, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	session, ok := s.sessions.Get(user)
	if !ok {
		return nil, errors.New("session not found")
	}

	collection, _ := s.managerCollection.Find(user, session.Collection)
	if collection != nil && collection.Status == domain.USER {
		return collection, nil
	}

	if collection != nil {
		collection.Status = domain.USER
	} else {
		collection = domain.NewUserCollection(user)
	}

	collection = s.managerCollection.Insert(user, collection)

	session.Collection = collection.Id

	s.update(session)

	return collection, nil
}

func (s *ManagerSession) FindUserHistoric(user string) (*domain.Collection, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	session, ok := s.sessions.Get(user)
	if !ok {
		return nil, errors.New("session not found")
	}

	collection, _ := s.managerCollection.Find(user, session.History)
	if collection != nil && collection.Status == domain.TALE {
		return collection, nil
	}

	if collection != nil {
		collection.Status = domain.TALE
	} else {
		collection = domain.NewUserCollection(user)
	}

	collection = s.managerCollection.Insert(user, collection)

	session.History = collection.Id

	s.update(session)

	return collection, nil
}

func (s *ManagerSession) FindUserGroup(user string) (*domain.Group, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	session, ok := s.sessions.Get(user)
	if !ok {
		return nil, errors.New("session not found")
	}

	group, _ := s.managerGroup.Find(user, session.Group)
	if group != nil {
		return group, nil
	}
	
	group = domain.NewGroup(user)
	group = s.managerGroup.Insert(user, group)

	session.Group = group.Id

	s.update(session)

	return group, nil
}

func (s *ManagerSession) Verify(username, oldPassword, newPassword1, newPassword2 string) (*session.Session, error) {
	err := s.valideData(username, newPassword1, &newPassword2)
	if err != nil {
		return nil, err
	}

	session, err := s.Authorize(username, oldPassword)
	if err != nil {
		return nil, errors.New("incorrect password")
	}

	if session == nil {
		return nil, errors.New("session not found")
	}

	secret, err := HashPassword(newPassword2)
	if err != nil {
		return nil, err
	}

	session.Secret = secret
	session.Count += 1
	s.update(session)

	return session, nil
}

func (s *ManagerSession) Delete(session *session.Session) (*session.Session, error) {
	if session.IsProtected {
		return nil, errors.New("this user is protected, cannot be removed")
	}

	session, _ = s.sessions.Remove(session.Username)
	return session, nil
}

func (s *ManagerSession) Visited(session *session.Session) *session.Session {
	session.Count += 1
	s.update(session)
	return session
}

func (s *ManagerSession) update(session *session.Session) (*session.Session, bool) {
	if _, ok := s.sessions.Get(session.Username); !ok {
		return nil, false
	}
	old, exists := s.sessions.Put(session.Username, *session)
	go s.write(s.sessions)
	return old, exists
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

func (s *ManagerSession) Insert(session *session.Session, user, password string, isAdmin bool) (*session.Session, error) {
	if !session.IsAdmin {
		return nil, errors.New("user has not have admin privilegies")
	}

	if _, exists := s.Find(user); exists {
		return nil, errors.New("user exists")
	}

	err := s.valideData(user, password, nil)
	if err != nil {
		return nil, err
	}

	collection, history, group := s.makeDependencies(user)
	return s.insert(user, password, collection, history, group, false, isAdmin, -1)
}

func (s *ManagerSession) insert(user, password string, collection *domain.Collection, history *domain.Collection, group *domain.Group, isProtected, isAdmin bool, count int) (*session.Session, error) {
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
		Username:    user,
		Secret:      secret,
		Timestamp:   time.Now().UnixMilli(),
		Collection:  collection.Id,
		History:     history.Id,
		Group:       group.Id,
		IsProtected: isProtected,
		IsAdmin:     isAdmin,
		Count:       count,
	}

	s.sessions.Put(user, session)

	go s.write(s.sessions)

	return &session, nil
}

func (s *ManagerSession) makeDependencies(username string) (*domain.Collection, *domain.Collection, *domain.Group) {
	collection := domain.NewUserCollection(username)
	collection.Name = fmt.Sprintf("%s's global collection", username)
	collection = s.managerCollection.Insert(username, collection)

	history := domain.NewTaleCollection(username)
	history.Name = fmt.Sprintf("%s's history collection", username)
	history.Context = collection.Context
	history = s.managerCollection.Insert(username, history)

	group := domain.NewGroup(username)
	group = s.managerGroup.Insert(username, group)

	return collection, history, group
}

func (s *ManagerSession) valideData(username, password1 string, password2 *string) error {
	if username == "" {
		return errors.New("invalid username")
	}

	if password1 == "" {
		return errors.New("invalid password")
	}

	if password2 == nil {
		return nil
	}

	if password1 != *password2 {
		return errors.New("new passwords doesn't matches")
	}

	return nil
}

func (s *ManagerSession) write(snapshot collection.IDictionary[string, session.Session]) {
	s.mutFile.Lock()
	defer s.mutFile.Unlock()

	items := collection.DictionaryMap(snapshot, func(k string, v session.Session) any {
		return *dto.FromSession(v)
	}).Values()

	err := s.file.Write(items)
	if err != nil {
		log.Error(err)
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
