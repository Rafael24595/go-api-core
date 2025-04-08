package repository

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type policy func(*domain.Request, IRepositoryRequest, IRepositoryResponse) error

const (
	POLICY_INSERT = "insert"
)

type ManagerRequest struct {
	mu       sync.Mutex
	request  IRepositoryRequest
	response IRepositoryResponse
	policies map[string][]policy
}

func NewManagerRequest(request IRepositoryRequest, response IRepositoryResponse) *ManagerRequest {
	return NewManagerRequestLimited(request, response)
}

func NewManagerRequestLimited(request IRepositoryRequest, response IRepositoryResponse) *ManagerRequest {
	return &ManagerRequest{
		request:  request,
		response: response,
		policies: make(map[string][]policy),
	}
}

func (m *ManagerRequest) SetInsertPolicy(function policy) *ManagerRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.policies[POLICY_INSERT]; !ok {
		m.policies[POLICY_INSERT] = []policy{}
	}

	m.policies[POLICY_INSERT] = append(m.policies[POLICY_INSERT], function)
	return m
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

func (m *ManagerRequest) FindOwner(owner string, status *domain.Status) []domain.Request {
	return m.request.FindOwner(owner, status)
}

func (m *ManagerRequest) FindSteps(steps []domain.Historic) []domain.Request {
	return m.request.FindSteps(steps)
}

func (m *ManagerRequest) FindNodes(nodes []domain.NodeReference) []dto.DtoNode {
	return m.request.FindNodes(nodes)
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

func (m *ManagerRequest) Insert(owner string, request *domain.Request, response *domain.Response) (*domain.Request, *domain.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := domain.StatusFromString(string(request.Status)); err != nil {
		request.Status = domain.DRAFT
	}

	requestResult := m.request.Insert(owner, request)

	response.Id = requestResult.Id
	response.Request = requestResult.Id
	resultResponse := m.response.Insert(owner, response)

	policies, ok := m.policies[POLICY_INSERT]
	if !ok {
		return requestResult, resultResponse
	}

	for _, p := range policies {
		p(requestResult, m.request, m.response)
	}

	return requestResult, resultResponse
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

func (m *ManagerRequest) Delete(owner string, request domain.Request) (*domain.Request, *domain.Response) {
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
