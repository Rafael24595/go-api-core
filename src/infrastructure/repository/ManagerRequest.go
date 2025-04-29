package repository

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
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

func (m *ManagerRequest) Exists(key string) (bool, bool) {
	_, okReq := m.request.Find(key)
	_, okRes := m.response.Find(key)
	return okReq, okRes
}

func (m *ManagerRequest) FindAll() []domain.Request {
	return m.request.FindAll()
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

func (m *ManagerRequest) FindNodes(nodes []domain.NodeReference) []dto.DtoNodeRequest {
	return m.request.FindNodes(nodes)
}

func (m *ManagerRequest) FindRequests(nodes []domain.NodeReference) []domain.Request {
	return m.request.FindRequests(nodes)
}


func (m *ManagerRequest) Release(owner string, request *domain.Request, response *domain.Response) (*domain.Request, *domain.Response) {
	if request.Status == domain.DRAFT {
		request.Status = domain.FINAL
		request.Id = ""
		request.Timestamp = time.Now().UnixMilli()
		request.Modified = request.Timestamp
	}
	return m.Insert(owner, request, response)
}

func (m *ManagerRequest) Insert(owner string, request *domain.Request, response *domain.Response) (*domain.Request, *domain.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()

	requestResult := m.request.Insert(owner, request)

	response.Id = requestResult.Id
	response.Request = requestResult.Id
	resultResponse := m.response.Insert(owner, response)

	return requestResult, resultResponse
}

func (m *ManagerRequest) InsertRequest(owner string, request *domain.Request) *domain.Request {
	return  m.request.Insert(owner, request)
}

func (m *ManagerRequest) InsertResponse(owner string, response *domain.Response) *domain.Response {
	return  m.response.Insert(owner, response)
}

func (m *ManagerRequest) InsertManyRequest(owner string, requests []domain.Request) []domain.Request {
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

func (m *ManagerRequest) Delete(owner string, request *domain.Request) (*domain.Request, *domain.Response) {
	return m.DeleteById(owner, request.Id)
}

func (m *ManagerRequest) DeleteById(owner, id string) (*domain.Request, *domain.Response) {
	request, exists := m.request.Find(id)
	if exists && request.Owner != owner {
		panic("//TODO: Manage error")	
	}

	request = m.request.DeleteById(id)
	response := m.response.DeleteById(id)
	return request, response
}

func (m *ManagerRequest) DeleteMany(owner string, ids ...string) ([]domain.Request, []domain.Response) {
	requests := m.request.DeleteMany(ids...)
	responses := m.response.DeleteMany(ids...)
	return requests, responses
}

func (m *ManagerRequest) DeleteManyRequests(owner string, ids ...string) []domain.Request {
	return m.request.DeleteMany(ids...)
}

func (m *ManagerRequest) DeleteManyResponses(owner string, ids ...string) []domain.Response {
	return m.response.DeleteMany(ids...)
}
