package repository

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
	"github.com/Rafael24595/go-collections/collection"
)

type ManagerRequest struct {
	mu       sync.Mutex
	request  IRepositoryRequest
	response IRepositoryResponse
}

func NewManagerRequest(request IRepositoryRequest, response IRepositoryResponse) *ManagerRequest {
	return NewManagerRequestLimited(request, response)
}

func NewManagerRequestLimited(request IRepositoryRequest, response IRepositoryResponse) *ManagerRequest {
	return &ManagerRequest{
		request:  request,
		response: response,
	}
}

func (m *ManagerRequest) Find(owner string, key string) (*domain.Request, *domain.Response, bool) {
	request, exits := m.request.Find(key)
	if !exits || request.Owner != owner  {
		return nil, nil, exits
	}
	response, _ := m.response.Find(key)

	return request, response, exits
}

func (m *ManagerRequest) FindRequest(owner string, key string) (*domain.Request, bool) {
	request, exits := m.request.Find(key)
	if !exits || request.Owner != owner  {
		return nil, exits
	}
	return request, exits
}

func (m *ManagerRequest) FindResponse(owner string, key string) (*domain.Response, bool) {
	response, exits := m.response.Find(key)
	if !exits || response.Owner != owner  {
		return nil, exits
	}
	return response, exits
}

func (m *ManagerRequest) FindLiteNodes(owner string, nodes []domain.NodeReference) []dto.DtoLiteNodeRequest {
	requests := m.request.FindLiteNodes(nodes)
	return collection.VectorFromList(requests).
		Filter(func(n dto.DtoLiteNodeRequest) bool {
			return n.Request.Owner == owner
		}).
		Collect()
}

func (m *ManagerRequest) FindNodes(owner string, nodes []domain.NodeReference) []dto.DtoNodeRequest {
	requests := m.request.FindNodes(nodes)
	return collection.VectorFromList(requests).
		Filter(func(n dto.DtoNodeRequest) bool {
			return n.Request.Owner == owner
		}).
		Collect()
}

func (m *ManagerRequest) FindRequests(owner string, nodes []domain.NodeReference) []domain.Request {
	requests := m.request.FindRequests(nodes)
	return collection.
		VectorFromList(requests).Filter(func(n domain.Request) bool {
			return n.Owner == owner
		}).
		Collect()
}

func (m *ManagerRequest) Release(owner string, request *domain.Request, response *domain.Response) (*domain.Request, *domain.Response) {
	if m.isNotOwner(owner, request, response) {
		return nil, nil
	}

	if request.Status == domain.DRAFT {
		request.Status = domain.FINAL
		request.Id = ""
		request.Timestamp = time.Now().UnixMilli()
		request.Modified = request.Timestamp
	}
	return m.Insert(owner, request, response)
}

func (m *ManagerRequest) Insert(owner string, request *domain.Request, response *domain.Response) (*domain.Request, *domain.Response) {
	if m.isNotOwner(owner, request, response) {
		return nil, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	requestResult := m.request.Insert(owner, request)

	response.Id = requestResult.Id
	response.Request = requestResult.Id
	resultResponse := m.response.Insert(owner, response)

	return requestResult, resultResponse
}

func (m *ManagerRequest) InsertRequest(owner string, request *domain.Request) *domain.Request {
	if m.isNotOwner(owner, request, nil) {
		return nil
	}
	return  m.request.Insert(owner, request)
}

func (m *ManagerRequest) InsertResponse(owner string, response *domain.Response) *domain.Response {
	if m.isNotOwner(owner, nil, response) {
		return nil
	}
	return  m.response.Insert(owner, response)
}

func (m *ManagerRequest) InsertManyRequest(owner string, requests []domain.Request) []domain.Request {
	requests = collection.VectorFromList(requests).
		Filter(func(r domain.Request) bool {
			return m.isOwner(owner, &r, nil)
		}).
		Collect()
	return m.request.InsertMany(owner, requests)
}

func (m *ManagerRequest) ImportDtoRequests(owner string, dtos []dto.DtoRequest) []domain.Request {
	m.mu.Lock()
	defer m.mu.Unlock()

	requests := make([]domain.Request, len(dtos))

	for i, v := range dtos {
		v.Id = ""
		request := dto.ToRequest(&v)
		request = m.request.Insert(owner, request)
		requests[i] = *request
	}
	
	return requests
}

func (m *ManagerRequest) Update(owner string, request *domain.Request) *domain.Request {
	oldRequest, exists := m.request.Find(request.Id)
	if !exists || oldRequest.Owner != owner {
		return request
	}

	if request.Status == domain.DRAFT {
		request.Name = oldRequest.Name
	}

	return m.request.Insert(owner, request)
}

func (m *ManagerRequest) DeleteById(owner, id string) (*domain.Request, *domain.Response) {
	request, exists := m.request.Find(id)
	if exists && request.Owner == owner {
		request = m.request.Delete(request)
	}

	response, exists := m.response.Find(id)
	if exists && response.Owner != owner {
		response = m.response.Delete(response)
	}

	return request, response
}

func (m *ManagerRequest) Delete(owner string, request *domain.Request) (*domain.Request, *domain.Response) {
	if request.Owner != owner {
		return nil, nil
	}
	return m.DeleteById(owner, request.Id)
}

func (m *ManagerRequest) DeleteMany(owner string, ids ...string) ([]domain.Request, []domain.Response) {
	requests := m.DeleteManyRequests(owner, ids...)
	responses := m.DeleteManyResponses(owner, ids...)
	return requests, responses
}

func (m *ManagerRequest) DeleteManyRequests(owner string, ids ...string) []domain.Request {
	requests := m.request.FindMany(ids)
	requests = collection.VectorFromList(requests).
		Filter(func(r domain.Request) bool {
			return r.Owner == owner
		}).
		Collect()
	return m.request.DeleteMany(requests...)
}

func (m *ManagerRequest) DeleteManyResponses(owner string, ids ...string) []domain.Response {
	responses := m.response.FindMany(ids)
	responses = collection.VectorFromList(responses).
		Filter(func(r domain.Response) bool {
			return r.Owner == owner
		}).
		Collect()
	return m.response.DeleteMany(responses...)
}

func (m *ManagerRequest) isNotOwner(owner string, request *domain.Request, response * domain.Response) bool {
	return !m.isOwner(owner, request, response)
}

func (m *ManagerRequest) isOwner(owner string, request *domain.Request, response * domain.Response) bool {
	if (request == nil && response == nil) {
		return false
	}

	if (request != nil && request.Id != "" && request.Owner != owner) {
		return false
	}

	if (response != nil && response.Id != "" && response.Owner != owner) {
		return false
	}
	
	return true
}
