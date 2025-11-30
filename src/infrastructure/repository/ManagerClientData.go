package repository

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/client"
	collection_domain "github.com/Rafael24595/go-api-core/src/domain/collection"
)

type ManagerClientData struct {
	mu                sync.Mutex
	client            IRepositoryClientData
	managerCollection *ManagerCollection
	managerGroup      *ManagerGroup
}

func NewManagerClientData(
	client IRepositoryClientData,
	managerCollection *ManagerCollection,
	managerGroup *ManagerGroup) *ManagerClientData {
	return &ManagerClientData{
		client:            client,
		managerCollection: managerCollection,
		managerGroup:      managerGroup,
	}
}

func (m *ManagerClientData) FindPersistent(user string) (*collection_domain.Collection, error) {
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

	collection, _ := m.managerCollection.Find(user, data.Persistent)
	if collection != nil && collection.Status == collection_domain.USER {
		return collection, nil
	}

	exists := collection != nil
	if exists {
		collection.Status = collection_domain.USER
	} else {
		collection = collection_domain.NewUserCollection(user)
	}

	collection = m.managerCollection.Insert(user, collection)

	if !exists {
		log.Messagef("Defined global collection '%s' for %s user", collection.Id, user)
	}

	data.Persistent = collection.Id

	m.client.Update(data)

	return collection, nil
}

func (m *ManagerClientData) FindTransient(user string) (*collection_domain.Collection, error) {
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

	collection, _ := m.managerCollection.Find(user, data.Transient)
	if collection != nil && collection.Status == collection_domain.TALE {
		return collection, nil
	}

	exists := collection != nil
	if exists {
		collection.Status = collection_domain.TALE
	} else {
		collection = collection_domain.NewUserCollection(user)
	}

	collection = m.managerCollection.Insert(user, collection)

	if !exists {
		log.Messagef("Defined historic collection '%s' for %s user", collection.Id, user)
	}

	data.Transient = collection.Id

	m.client.Update(data)

	return collection, nil
}

func (m *ManagerClientData) FindCollections(user string) (*domain.Group, error) {
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

	group, _ := m.managerGroup.Find(user, data.Collections)
	if group != nil {
		return group, nil
	}

	group = domain.NewGroup(user)
	group = m.managerGroup.Insert(user, group)

	data.Collections = group.Id

	m.client.Update(data)

	return group, nil
}

func (m *ManagerClientData) valideSessionAndRelease(user string) (*client.ClientData, error) {
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

func (m *ManagerClientData) resolve(owner string) (*client.ClientData, bool) {
	if data, exists := m.client.Find(owner); exists {
		return data, true
	}

	data := m.releaseClientData(owner)
	return m.client.Insert(data), false
}

func (m *ManagerClientData) releaseClientData(owner string) *client.ClientData {
	collection := collection_domain.NewUserCollection(owner)
	collection.Name = fmt.Sprintf("%s's global collection", owner)
	collection = m.managerCollection.Insert(owner, collection)

	history := collection_domain.NewTaleCollection(owner)
	history.Name = fmt.Sprintf("%s's history collection", owner)
	history.Context = collection.Context
	history = m.managerCollection.Insert(owner, history)

	group := domain.NewGroup(owner)
	group = m.managerGroup.Insert(owner, group)

	return client.NewClientData(owner, history.Id, collection.Id, group.Id)
}

func (m *ManagerClientData) delete(owner string) *client.ClientData {
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
