package session

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	domain_session "github.com/Rafael24595/go-api-core/src/domain/session"
	"github.com/Rafael24595/go-collections/collection"
	"golang.org/x/crypto/bcrypt"
)

const NameMemory = "request_memory"

var (
	instance *ManagerSession
	once     sync.Once
)

type ManagerSession struct {
	mut               sync.RWMutex
	sessions          domain_session.RepositorySession
	managerSessionData *ManagerSessionData
}

func InitializeManagerSession(conf configuration.Configuration, sessions domain_session.RepositorySession, managerSessionData *ManagerSessionData) *ManagerSession {
	once.Do(func() {
		raw := &ManagerSession{
			sessions:          sessions,
			managerSessionData: managerSessionData,
		}

		instance = resolveDefaultSessions(conf, raw)
	})

	if instance == nil {
		log.Panics("The session manager is not initialized properly")
	}

	return instance
}

func InstanceManagerSession() *ManagerSession {
	if instance == nil {
		log.Panics("The session manager is not initialized yet")
	}
	return instance
}

func resolveDefaultSessions(conf configuration.Configuration, instance *ManagerSession) *ManagerSession {
	rolesAdmin := []domain_session.Role{domain_session.ROLE_ADMIN, domain_session.ROLE_PROTECTED}
	err := instance.resolveDefaultSession(conf.Admin(), string(conf.Secret()), rolesAdmin, 0)
	if err != nil {
		log.Panic(err)
	}

	rolesAnonymous := []domain_session.Role{domain_session.ROLE_ANONYMOUS, domain_session.ROLE_PROTECTED}
	err = instance.resolveDefaultSession("anonymous", "", rolesAnonymous, 0)
	if err != nil {
		log.Panic(err)
	}

	return instance
}

func (s *ManagerSession) resolveDefaultSession(username, secret string, roles []domain_session.Role, count int) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	sess, exists := s.sessions.Find(username)
	if exists {
		if len(sess.Roles) != len(roles) {
			log.Messagef("Updating user %s with roles: %v", username, roles)
			sess.Roles = roles

			_, err := s.insert(sess)

			return err
		}

		if sess.Lock {
			log.Messagef("Unlocking user %s: %v", username)
			sess.Lock = false

			_, err := s.insert(sess)

			return err
		}
	}

	if !exists {
		log.Messagef("Defining default user %s with roles: %v", username, roles)

		_, err := s.insertWithContext(domain_session.SYSTEM_USER, username, secret, roles, count)

		return err
	}

	return nil
}

func (s *ManagerSession) GetPublicRoles() []domain_session.Role {
	return domain_session.PUBLIC_ROLES
}

func (s *ManagerSession) FindAll() []domain_session.SessionLite {
	return collection.MapToVector(s.sessions.FindAll(), func(s domain_session.Session) domain_session.SessionLite {
		return *domain_session.ToLite(s)
	}).Collect()
}

func (s *ManagerSession) Find(user string) (*domain_session.Session, bool) {
	return s.sessions.Find(user)
}

func (s *ManagerSession) FindSafe(user string) (*domain_session.SessionSafe, bool) {
	sess, ok := s.sessions.Find(user)
	if !ok {
		return nil, ok
	}
	return domain_session.ToSafe(*sess), true
}

func (s *ManagerSession) FindProvider(user string) (*domain_session.Session, error) {
	return s.valideProvider(user)
}

func (s *ManagerSession) Insert(provider *domain_session.Session, user, password string, roles []domain_session.Role) (*domain_session.Session, error) {
	provider, err := s.valideProvider(provider.Username)
	if err != nil {
		return nil, err
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	if _, exists := s.Find(user); exists {
		return nil, errors.New("user exists")
	}

	err = valideData(user, password, nil)
	if err != nil {
		return nil, err
	}

	roles = domain_session.CleanPrivateRoles(roles)

	return s.insertWithContext(provider.Username, user, password, roles, -1)
}

func (s *ManagerSession) Delete(provider, sess *domain_session.Session) (*domain_session.Session, error) {
	provider, err := s.valideProvider(provider.Username)
	if err != nil {
		return nil, err
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	sess, exists := s.Find(sess.Username)
	if !exists {
		return nil, errors.New("user does not exists")
	}

	if sess.HasRole(domain_session.ROLE_PROTECTED) {
		return nil, errors.New("this user is protected, cannot be removed")
	}

	s.managerSessionData.delete(sess.Username)

	return s.sessions.Delete(sess), nil
}

func (s *ManagerSession) Verify(username, oldPassword, newPassword1, newPassword2 string) (*domain_session.Session, error) {
	err := valideData(username, newPassword1, &newPassword2)
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

	secret, err := hashPassword(newPassword2)
	if err != nil {
		return nil, err
	}

	session.Secret = secret

	return s.Visited(session), nil
}

func (s *ManagerSession) Authorize(user, password string) (*domain_session.Session, error) {
	session, exists := s.sessions.Find(user)
	if !exists {
		return nil, errors.New("session not found")
	}

	if !valideSecret([]byte(password), session.Secret) {
		return nil, errors.New("session not found")
	}

	return session, nil
}

func (s *ManagerSession) Refresh(session *domain_session.Session, refresh string) *domain_session.Session {
	session, exists := s.sessions.Find(session.Username)
	if !exists {
		return nil
	}

	session.Refresh = refresh
	s.update(session)
	return session
}

func (s *ManagerSession) Visited(session *domain_session.Session) *domain_session.Session {
	session, exists := s.sessions.Find(session.Username)
	if !exists {
		return nil
	}

	session.Count += 1
	s.update(session)
	return session
}

func (s *ManagerSession) Lock(provider, session *domain_session.Session) (*domain_session.Session, error) {
	return s.updateStatus(provider, session, false)
}

func (s *ManagerSession) Unlock(provider, session *domain_session.Session) (*domain_session.Session, error) {
	return s.updateStatus(provider, session, true)
}

func (s *ManagerSession) updateStatus(provider, session *domain_session.Session, status bool) (*domain_session.Session, error) {
	provider, err := s.valideProvider(provider.Username)
	if err != nil {
		return nil, err
	}

	session, exists := s.sessions.Find(session.Username)
	if !exists {
		return nil, err
	}

	session.Lock = status
	s.update(session)

	return session, nil
}

func (s *ManagerSession) valideProvider(user string) (*domain_session.Session, error) {
	provider, exists := s.Find(user)
	if !exists || !provider.HasRole(domain_session.ROLE_ADMIN) || provider.Lock {
		return nil, fmt.Errorf("Access is denied. User '%s' does not have sufficient privileges", user)
	}
	return provider, nil
}

func (s *ManagerSession) insertWithContext(provider string, user, password string, roles []domain_session.Role, count int) (*domain_session.Session, error) {
	_, exists := s.sessions.Find(user)
	if exists {
		return nil, errors.New("user already exists")
	}

	data, exists := s.managerSessionData.resolve(user)
	if data == nil && !exists {
		return nil, errors.New("client data cannot be initialized")
	}

	sess, err := s.makeUser(provider, user, password, roles, count)
	if err != nil {
		s.managerSessionData.delete(user)
		return nil, err
	}

	sess, err = s.insert(sess)
	if err != nil {
		s.managerSessionData.delete(user)
		return nil, err
	}

	return sess, nil
}

func (s *ManagerSession) insert(sess *domain_session.Session) (*domain_session.Session, error) {
	sess.Roles = domain_session.Unique(sess.Roles)

	if sess.Lock && sess.HasRole(domain_session.ROLE_PROTECTED) {
		sess.Lock = false
	}

	if sess.Timestamp == 0 {
		sess.Timestamp = time.Now().UnixMilli()
	}

	return s.sessions.Insert(sess), nil
}

func (s *ManagerSession) update(session *domain_session.Session) (*domain_session.Session, error) {
	if _, ok := s.sessions.Find(session.Username); !ok {
		return nil, nil
	}

	return s.insert(session)
}

func (s *ManagerSession) makeUser(publisher, user, password string, roles []domain_session.Role, count int) (*domain_session.Session, error) {
	secret, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	return &domain_session.Session{
		Username:  user,
		Lock:      false,
		Secret:    secret,
		Timestamp: time.Now().UnixMilli(),
		Publisher: publisher,
		Count:     count,
		Refresh:   "",
		Roles:     roles,
	}, nil
}

func valideData(username, password1 string, password2 *string) error {
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

func hashPassword(password string) ([]byte, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return hashedBytes, err
}

func valideSecret(password, hashed []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashed, password)
	return err == nil
}
