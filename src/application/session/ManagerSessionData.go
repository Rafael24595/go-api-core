package session

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Rafael24595/go-api-core/src/application/manager"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/domain/group"
	"github.com/Rafael24595/go-api-core/src/domain/session"
)

type ManagerSessionData struct {
	mu                sync.Mutex
	client            session.RepositorySessionData
	managerCollection *manager.ManagerCollection
	managerGroup      *manager.ManagerGroup
}

func NewManagerSessionData(
	client session.RepositorySessionData,
	managerCollection *manager.ManagerCollection,
	managerGroup *manager.ManagerGroup,
) *ManagerSessionData {
	return &ManagerSessionData{
		client:            client,
		managerCollection: managerCollection,
		managerGroup:      managerGroup,
	}
}

func (m *ManagerSessionData) FindPersistent(user string) (*collection.Collection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.client.Find(user)
	if !ok {
		result, err := m.valideSessionAndRelease(user)
		if err != nil {
			return nil, err
		}

		data = result
	}

	coll, _ := m.managerCollection.Find(user, data.Persistent)
	if coll != nil && coll.Status == collection.USER {
		return coll, nil
	}

	exists := coll != nil
	if exists {
		coll.Status = collection.USER
	} else {
		coll = collection.NewUserCollection(user)
	}

	coll = m.managerCollection.Insert(user, coll)

	if !exists {
		log.Messagef("Defined global collection '%s' for %s user", coll.Id, user)
	}

	data.Persistent = coll.Id

	m.client.Update(data)

	return coll, nil
}

func (m *ManagerSessionData) FindTransient(user string) (*collection.Collection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.client.Find(user)
	if !ok {
		result, err := m.valideSessionAndRelease(user)
		if err != nil {
			return nil, err
		}

		data = result
	}

	coll, _ := m.managerCollection.Find(user, data.Transient)
	if coll != nil && coll.Status == collection.TALE {
		return coll, nil
	}

	exists := coll != nil
	if exists {
		coll.Status = collection.TALE
	} else {
		coll = collection.NewUserCollection(user)
	}

	coll = m.managerCollection.Insert(user, coll)

	if !exists {
		log.Messagef("Defined historic collection '%s' for %s user", coll.Id, user)
	}

	data.Transient = coll.Id

	m.client.Update(data)

	return coll, nil
}

func (m *ManagerSessionData) FindCollections(user string) (*group.Group, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.client.Find(user)
	if !ok {
		result, err := m.valideSessionAndRelease(user)
		if err != nil {
			return nil, err
		}

		data = result
	}

	grp, _ := m.managerGroup.Find(user, data.Collections)
	if grp != nil {
		return grp, nil
	}

	grp = group.NewGroup(user)
	grp = m.managerGroup.Insert(user, grp)

	data.Collections = grp.Id

	m.client.Update(data)

	return grp, nil
}

func (m *ManagerSessionData) valideSessionAndRelease(user string) (*session.ClientData, error) {
	_, ok := InstanceManagerSession().Find(user)
	if !ok {
		return nil, errors.New("user does not exists")
	}

	result, ok := m.resolve(user)
	if !ok {
		return nil, fmt.Errorf("cannot generate client data for user %q", user)
	}

	return result, nil
}

func (m *ManagerSessionData) resolve(owner string) (*session.ClientData, bool) {
	if data, exists := m.client.Find(owner); exists {
		return data, true
	}

	data := m.releaseClientData(owner)
	return m.client.Insert(data), false
}

func (m *ManagerSessionData) releaseClientData(owner string) *session.ClientData {
	coll := collection.NewUserCollection(owner)
	coll.Name = fmt.Sprintf("%s's global collection", owner)
	coll = m.managerCollection.Insert(owner, coll)

	history := collection.NewTaleCollection(owner)
	history.Name = fmt.Sprintf("%s's history collection", owner)
	history.Context = coll.Context
	history = m.managerCollection.Insert(owner, history)

	group := group.NewGroup(owner)
	group = m.managerGroup.Insert(owner, group)

	return session.NewClientData(owner, history.Id, coll.Id, group.Id)
}

func (m *ManagerSessionData) delete(owner string) *session.ClientData {
	data, exists := m.client.Find(owner)
	if !exists {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.managerCollection.Delete(owner, data.Transient)
	m.managerCollection.Delete(owner, data.Persistent)
	m.managerGroup.Delete(owner, data.Collections)

	return m.client.Delete(data)
}
