package repository

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/context"
)

type ManagerCollection struct {
	mu         sync.Mutex
	collection IRepositoryCollection
	context    IRepositoryContext
	request    IRepositoryRequest
	response   IRepositoryResponse
}

func NewManagerCollection(collection IRepositoryCollection, context IRepositoryContext, request IRepositoryRequest, response IRepositoryResponse) *ManagerCollection {
	return &ManagerCollection{
		collection: collection,
		context:    context,
		request:    request,
		response:   response,
	}
}

func (m *ManagerCollection) Exists(key string) bool {
	return m.collection.Exists(key)
}

func (m *ManagerCollection) FindByOwner(owner string) []domain.Collection {
	return m.collection.FindByOwner(owner)
}

func (m *ManagerCollection) Insert(owner string, collection *domain.Collection) *domain.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.context.FindByCollection(owner, collection.Id); !exists {
		context := m.context.InsertFromCollection(owner, collection.Id, context.NewContext(owner))
		collection.Context = context.Id
	}

	return m.collection.Insert(owner, collection)
}

func (m *ManagerCollection) PushToCollection(owner string, collectionId string, collectionName string, request *domain.Request, requestName string) *domain.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	request.Name = requestName
	request.Status = domain.GROUP
	request = m.request.Insert(owner, request)

	return m.collection.PushToCollection(owner, collectionId, collectionName, request)
}

func (m *ManagerCollection) Delete(collection domain.Collection) *domain.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	if context, exists := m.context.FindByCollection(collection.Owner, collection.Id); !exists {
		m.context.Delete(*context)
	}

	requests := make([]string, len(collection.Nodes))
	for i, v := range collection.Nodes {
		requests[i] = v.Request
	}

	m.request.DeleteMany(requests...)
	m.response.DeleteMany(requests...)

	return m.collection.Delete(collection)
}
