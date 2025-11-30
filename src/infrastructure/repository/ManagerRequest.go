package repository

import (
	"sync"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
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

func (m *ManagerRequest) Export(owner string, nodes ...domain.NodeReference) []action.Request {
	ids := collection.MapToVector(nodes, func(n domain.NodeReference) string {
		return n.Item
	}).Collect()
	return m.ExportList(owner, ids...)
}

func (m *ManagerRequest) ExportList(owner string, ids ...string) []action.Request {
	requests := m.request.FindMany(ids...)
	return collection.VectorFromList(requests).
		Filter(func(n action.Request) bool {
			return n.Owner == owner
		}).
		Collect()
}

func (m *ManagerRequest) Find(owner string, key string) (*action.Request, *action.Response, bool) {
	request, exits := m.request.Find(key)
	if !exits || request.Owner != owner {
		return nil, nil, exits
	}
	response, _ := m.response.Find(key)

	return request, response, exits
}

func (m *ManagerRequest) FindRequest(owner string, key string) (*action.Request, bool) {
	request, exits := m.request.Find(key)
	if !exits || request.Owner != owner {
		return nil, exits
	}
	return request, exits
}

func (m *ManagerRequest) FindResponse(owner string, key string) (*action.Response, bool) {
	response, exits := m.response.Find(key)
	if !exits || response.Owner != owner {
		return nil, exits
	}
	return response, exits
}

func (m *ManagerRequest) FindLiteNodes(owner string, references []domain.NodeReference) []action.NodeRequestLite {
	nodes := m.request.FindNodes(references)
	lite := action.ToNodeRequestLite(nodes)
	return collection.VectorFromList(lite).
		Filter(func(n action.NodeRequestLite) bool {
			return n.Request.Owner == owner
		}).
		Collect()
}

func (m *ManagerRequest) FindNodes(owner string, nodes []domain.NodeReference) []action.NodeRequest {
	requests := m.request.FindNodes(nodes)
	return collection.VectorFromList(requests).
		Filter(func(n action.NodeRequest) bool {
			return n.Request.Owner == owner
		}).
		Collect()
}

func (m *ManagerRequest) Release(owner string, request *action.Request, response *action.Response) (*action.Request, *action.Response) {
	if m.isNotOwner(owner, request, response) {
		return nil, nil
	}

	if request.Status == action.DRAFT {
		request.Status = action.FINAL
		request.Id = ""
		request.Timestamp = time.Now().UnixMilli()
		request.Modified = request.Timestamp
	}
	return m.Insert(owner, request, response)
}

func (m *ManagerRequest) Insert(owner string, request *action.Request, response *action.Response) (*action.Request, *action.Response) {
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

func (m *ManagerRequest) InsertRequest(owner string, request *action.Request) *action.Request {
	if m.isNotOwner(owner, request, nil) {
		return nil
	}
	return m.request.Insert(owner, request)
}

func (m *ManagerRequest) InsertResponse(owner string, response *action.Response) *action.Response {
	if m.isNotOwner(owner, nil, response) {
		return nil
	}
	return m.response.Insert(owner, response)
}

func (m *ManagerRequest) InsertManyRequests(owner string, requests []action.Request) []action.Request {
	requests = collection.VectorFromList(requests).
		Filter(func(r action.Request) bool {
			return m.isOwner(owner, &r, nil)
		}).
		Collect()
	return m.request.InsertMany(owner, requests)
}

func (m *ManagerRequest) Update(owner string, request *action.Request) *action.Request {
	oldRequest, exists := m.request.Find(request.Id)
	if !exists || oldRequest.Owner != owner {
		return request
	}

	if request.Status == action.DRAFT {
		request.Name = oldRequest.Name
	}

	return m.request.Insert(owner, request)
}

func (m *ManagerRequest) Delete(owner string, request *action.Request) (*action.Request, *action.Response) {
	if request.Owner != owner {
		return nil, nil
	}
	return m.deleteById(owner, request.Id)
}

func (m *ManagerRequest) deleteById(owner, id string) (*action.Request, *action.Response) {
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

func (m *ManagerRequest) DeleteMany(owner string, ids ...string) ([]action.Request, []action.Response) {
	requests := m.deleteManyRequests(owner, ids...)
	responses := m.deleteManyResponses(owner, ids...)
	return requests, responses
}

func (m *ManagerRequest) deleteManyRequests(owner string, ids ...string) []action.Request {
	requests := m.request.FindMany(ids...)
	requests = collection.VectorFromList(requests).
		Filter(func(r action.Request) bool {
			return r.Owner == owner
		}).
		Collect()
	return m.request.DeleteMany(requests...)
}

func (m *ManagerRequest) deleteManyResponses(owner string, ids ...string) []action.Response {
	responses := m.response.FindMany(ids)
	responses = collection.VectorFromList(responses).
		Filter(func(r action.Response) bool {
			return r.Owner == owner
		}).
		Collect()
	return m.response.DeleteMany(responses...)
}

func (m *ManagerRequest) isNotOwner(owner string, request *action.Request, response *action.Response) bool {
	return !m.isOwner(owner, request, response)
}

func (m *ManagerRequest) isOwner(owner string, request *action.Request, response *action.Response) bool {
	if request == nil && response == nil {
		return false
	}

	if request != nil && request.Id != "" && request.Owner != owner {
		return false
	}

	if response != nil && response.Id != "" && response.Owner != owner {
		return false
	}

	return true
}
