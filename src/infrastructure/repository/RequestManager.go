package repository

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
)

type policy func(*domain.Request, IRepositoryRequest, IRepositoryResponse) error

const (
	POLICY_INSERT = "insert"
)

type RequestManager struct {
	mu       sync.Mutex
	request  IRepositoryRequest
	response IRepositoryResponse
	policies map[string][]policy
}

func NewRequestManager(request IRepositoryRequest, response IRepositoryResponse) *RequestManager {
	return NewRequestManagerLimited(request, response)
}

func NewRequestManagerLimited(request IRepositoryRequest, response IRepositoryResponse) *RequestManager {
	return &RequestManager{
		request:  request,
		response: response,
		policies: make(map[string][]policy),
	}
}

func (m *RequestManager) SetInsertPolicy(function policy) *RequestManager {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.policies[POLICY_INSERT]; !ok {
		m.policies[POLICY_INSERT] = []policy{}
	}

	m.policies[POLICY_INSERT] = append(m.policies[POLICY_INSERT], function)
	return m
}

func (m *RequestManager) Exists(key string) (bool, bool) {
	_, okReq := m.request.Find(key)
	_, okRes := m.response.Find(key)
	return okReq, okRes
}

func (m *RequestManager) FindAll() []domain.Request {
	return m.request.FindAll()
}

func (m *RequestManager) Find(key string) (*domain.Request, *domain.Response, bool) {
	request, ok := m.request.Find(key)
	if !ok {
		return nil, nil, ok
	}
	response, _ := m.response.Find(key)
	return request, response, ok
}

func (m *RequestManager) FindOptions(options FilterOptions[domain.Request]) []domain.Request {
	return m.request.FindOptions(options)
}

func (m *RequestManager) FindSteps(steps []domain.Historic) []domain.Request {
	return m.request.FindSteps(steps)
}

func (m *RequestManager) Insert(request domain.Request, response *domain.Response) (*domain.Request, *domain.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()

	requestResult := m.request.Insert(request)

	response.Id = requestResult.Id
	response.Request = requestResult.Id
	resultResponse := m.response.Insert(*response)

	policies, ok := m.policies[POLICY_INSERT]
	if !ok {
		return &requestResult, &resultResponse
	}

	for _, p := range policies {
		p(&requestResult, m.request, m.response)
	}

	return &requestResult, &resultResponse
}

func (m *RequestManager) Delete(request domain.Request) (*domain.Request, *domain.Response) {
	requestResult := m.request.Delete(request)
	response, ok := m.response.Find(request.Id)
	if ok {
		m.response.Delete(*response)
	}
	return requestResult, response
}
