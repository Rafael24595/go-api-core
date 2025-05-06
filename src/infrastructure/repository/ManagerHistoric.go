package repository

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type ManagerHistoric struct {
	mu                sync.Mutex
	managerRequest    *ManagerRequest
	managerCollection *ManagerCollection
}

func NewManagerHistoric(managerRequest *ManagerRequest, managerCollection *ManagerCollection) *ManagerHistoric {
	return &ManagerHistoric{
		managerRequest:    managerRequest,
		managerCollection: managerCollection,
	}
}

func (m *ManagerHistoric) Find(owner string, collection *domain.Collection) []dto.DtoNodeRequest {
	return m.managerCollection.FindRequestNodes(owner, collection)
}

func (m *ManagerHistoric) Insert(owner string, collection *domain.Collection, request *domain.Request, response *domain.Response) (*domain.Collection, *domain.Request, *domain.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()

	request, response = m.managerRequest.Insert(owner, request, response)
	collection = m.managerCollection.ResolveRequestReferences(owner, collection, *request)

	if len(collection.Nodes) <= 10 {
		return collection, request, response
	}

	collection = collection.SortRequests()

	requests := make([]string, 0)
	nodes := make([]domain.NodeReference, 0)

	count := 0
	for i := len(collection.Nodes) - 1; i >= 0; i-- {
		v := collection.Nodes[i]
		if count > 9 {
			requests = append(requests, v.Item)
			continue
		}

		nodes = append(nodes, v)

		count++
	}

	m.managerRequest.DeleteMany(owner, requests...)

	collection.Nodes = nodes
	collection = collection.SortRequests().FixRequestsOrder()
	collection = m.managerCollection.Insert(owner, collection)

	return collection, request, response
}

func (m *ManagerHistoric) Delete(owner string, collection *domain.Collection, requestId string) (*domain.Collection, *domain.Request, *domain.Response) {
	return m.managerCollection.DeleteRequestFromCollection(owner, collection, requestId)
}
