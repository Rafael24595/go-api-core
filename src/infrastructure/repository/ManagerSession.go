package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/session"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/system/topic"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-collections/collection"
	"golang.org/x/crypto/bcrypt"
)

var (
	manager *ManagerSession
	once    sync.Once
)

type ManagerSession struct {
	once              sync.Once
	mut               sync.RWMutex
	mutFile           sync.RWMutex
	file              IFileManager[dto.DtoSession]
	sessions          collection.IDictionary[string, session.Session]
	managerClientData *ManagerClientData
	close             chan bool
}

func InitializeManagerSession(file IFileManager[dto.DtoSession], managerClientData *ManagerClientData) *ManagerSession {
	once.Do(func() {
		steps, err := file.Read()
		if err != nil {
			log.Panic(err)
			return
		}

		sessions := collection.MapToDictionarySync(steps,
			func(k string, d dto.DtoSession) session.Session {
				return *dto.ToSession(d)
			})

		instance := &ManagerSession{
			file:              file,
			managerClientData: managerClientData,
			sessions:          sessions,
		}

		manager = defineDefaultSessions(instance)

		go manager.watch()
	})

	if manager == nil {
		log.Panics("The session manager is not initialized properly")
	}

	return manager
}

func defineDefaultSessions(instance *ManagerSession) *ManagerSession {
	conf := configuration.Instance()

	rolesAdmin := []session.Role{session.ROLE_ADMIN, session.ROLE_PROTECTED}
	err := instance.defineDefaultUser(conf.Admin(), string(conf.Secret()), rolesAdmin, 0)
	if err != nil {
		log.Panic(err)
	}

	rolesAnonymous := []session.Role{session.ROLE_ANONYMOUS, session.ROLE_PROTECTED}
	err = instance.defineDefaultUser("anonymous", "", rolesAnonymous, 0)
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

func (r *ManagerSession) watch() {
	r.once.Do(func() {
		conf := configuration.Instance()
		if !conf.Snapshot().Enable {
			return
		}

		hub := make(chan system.SystemEvent, 1)
		defer close(hub)

		topics := []topic.TopicAction{
			topic_repository.TOPIC_SESSION.ActionReload(),
		}

		conf.EventHub.Subcribe(RepositoryListener, hub, topics...)
		defer conf.EventHub.Unsubcribe(RepositoryListener, topics...)

		for {
			select {
			case <-r.close:
				log.Customf(SnapshotCategory, "Watcher stopped: local close signal received.")
				return
			case <-hub:
				if err := r.read(); err != nil {
					log.Custome(SnapshotCategory, err)
					return
				}
			case <-conf.Signal.Done():
				log.Customf(SnapshotCategory, "Watcher stopped: global shutdown signal received.")
				return
			}
		}
	})
}

func (r *ManagerSession) read() error {
	sessions, err := r.file.Read()
	if err != nil {
		return err
	}

	r.sessions = collection.DictionarySyncMap(
		collection.DictionarySyncFromMap(sessions),
		func(k string, d dto.DtoSession) session.Session {
			return *dto.ToSession(d)
		})

	return nil
}

func (s *ManagerSession) defineDefaultUser(username, secret string, roles []session.Role, count int) error {
	sess, exists := s.sessions.Get(username)
	if exists && len(sess.Roles) != len(roles) {
		log.Messagef("Updating user %s with roles: %v", username, roles)
		sess.Roles = roles
		s.sessions.Put(sess.Username, sess)

		go s.write(s.sessions)

		return nil
	}

	if !exists {
		log.Messagef("Defining default user %s with roles: %v", username, roles)

		data, exists := s.managerClientData.resolve(username)
		if data == nil && !exists {
			return errors.New("client data cannot be initialized")
		}

		_, err := s.insert(session.SYSTEM_USER, username, string(secret), roles, count)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ManagerSession) GetPublicRoles() []session.Role {
	return session.PUBLIC_ROLES
}

func (s *ManagerSession) FindAll() []session.SessionLite {
	return collection.MapToVector(s.sessions.Values(), func(s session.Session) session.SessionLite {
		return session.ToLite(s)
	}).Collect()
}

func (s *ManagerSession) FindSafe(user string) (*session.SessionSafe, bool) {
	sess, ok := s.sessions.Get(user)
	safe := session.ToSafe(sess)
	return &safe, ok
}

func (s *ManagerSession) Find(user string) (*session.Session, bool) {
	session, ok := s.sessions.Get(user)
	return &session, ok
}

func (s *ManagerSession) Insert(sess *session.Session, user, password string, roles []session.Role) (*session.Session, error) {
	if !sess.HasRole(session.ROLE_ADMIN) {
		return nil, errors.New("user has not have admin privilegies")
	}

	if _, exists := s.Find(user); exists {
		return nil, errors.New("user exists")
	}

	err := s.valideData(user, password, nil)
	if err != nil {
		return nil, err
	}

	data, exists := s.managerClientData.resolve(user)
	if data == nil && !exists {
		return nil, errors.New("client data cannot be initialized")
	}

	roles = manager.cleanPrivateRoles(roles)

	return s.insert(sess.Username, user, password, roles, -1)
}

func (s *ManagerSession) cleanPrivateRoles(roles []session.Role) []session.Role {
	cache := make(map[session.Role]bool, 0)
	fix := make([]session.Role, 0)

	for _, v := range roles {
		if session.IsPrivateRole(v) {
			continue
		}

		if _, ok := cache[v]; !ok {
			fix = append(fix, v)
			cache[v] = true
		}
	}

	return fix
}

func (s *ManagerSession) Delete(sess *session.Session) (*session.Session, error) {
	if sess.HasRole(session.ROLE_PROTECTED) {
		return nil, errors.New("this user is protected, cannot be removed")
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	s.managerClientData.delete(sess.Username)

	deleted, _ := s.sessions.Remove(sess.Username)
	return &deleted, nil
}

func (s *ManagerSession) Authorize(user, password string) (*session.Session, error) {
	session, exists := s.sessions.Get(user)
	if !exists {
		return nil, errors.New("session not found")
	}

	if !ValideSecret([]byte(password), session.Secret) {
		return nil, errors.New("session not found")
	}

	return &session, nil
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

func (s *ManagerSession) Visited(session *session.Session) *session.Session {
	session.Count += 1
	s.update(session)
	return session
}

func (s *ManagerSession) insert(publisher, user, password string, roles []session.Role, count int) (*session.Session, error) {
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

	roles = s.fixRoles(roles)

	session := session.Session{
		Username:  user,
		Secret:    secret,
		Timestamp: time.Now().UnixMilli(),
		Publisher: publisher,
		Count:     count,
		Refresh:   "",
		Roles:     roles,
	}

	s.sessions.Put(user, session)

	go s.write(s.sessions)

	return &session, nil
}

func (s *ManagerSession) fixRoles(roles []session.Role) []session.Role {
	cache := make(map[session.Role]bool, 0)
	fix := make([]session.Role, 0)

	for _, v := range roles {
		if _, ok := cache[v]; !ok {
			fix = append(fix, v)
			cache[v] = true
		}
	}

	return fix
}

func (s *ManagerSession) update(session *session.Session) (*session.Session, bool) {
	if _, ok := s.sessions.Get(session.Username); !ok {
		return nil, false
	}

	old, exists := s.sessions.Put(session.Username, *session)

	go s.write(s.sessions)

	return &old, exists
}

func (s *ManagerSession) Refresh(session *session.Session, refresh string) *session.Session {
	session.Refresh = refresh
	s.update(session)
	return session
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

	items := collection.DictionaryMap(snapshot, func(k string, v session.Session) dto.DtoSession {
		return *dto.FromSession(v)
	})

	err := s.file.Write(items.Values())
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
