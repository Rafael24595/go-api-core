package repository

import (
	"errors"
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/domain/openapi"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type ManagerCollection struct {
	mu             sync.Mutex
	collection     IRepositoryCollection
	managerContext *ManagerContext
	managerRequest *ManagerRequest
}

func NewManagerCollection(collection IRepositoryCollection, managerContext *ManagerContext, managerRequest *ManagerRequest) *ManagerCollection {
	return &ManagerCollection{
		collection:     collection,
		managerContext: managerContext,
		managerRequest: managerRequest,
	}
}

func (m *ManagerCollection) Exists(key string) bool {
	return m.collection.Exists(key)
}

func (m *ManagerCollection) Find(owner string, id string) (*domain.Collection, bool) {
	return m.collection.Find(id)
}

func (m *ManagerCollection) FindFreeByOwner(owner string) []domain.Collection {
	return m.collection.FindAllBystatus(owner, domain.FREE)
}

func (m *ManagerCollection) FindNodes(owner string, collection *domain.Collection) []dto.DtoNode {
	dtos := m.managerRequest.FindNodes(collection.Nodes)

	if len(dtos) == len(collection.Nodes) {
		return dtos
	}

	nodes := make([]domain.NodeReference, len(dtos))
	for _, v := range dtos {
		nodes = append(nodes, domain.NodeReference{
			Order:   v.Order,
			Request: v.Request.Id,
		})
	}

	collection.Nodes = nodes

	m.Insert(owner, collection)

	return dtos
}

func (m *ManagerCollection) Insert(owner string, collection *domain.Collection) *domain.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	if collection.Id == "" {
		collection = m.collection.Insert(owner, collection)
	}

	if _, exists := m.managerContext.Find(owner, collection.Context); !exists {
		context := m.managerContext.Insert(owner, collection.Id, context.NewContext(owner))
		collection.Context = context.Id
	}

	collection = collection.SortRequests().FixRequestsOrder()

	return m.collection.Insert(owner, collection)
}

func (m *ManagerCollection) ImportDtoCollections(owner string, dtos []dto.DtoCollection) ([]domain.Collection, error) {
	collections := make([]domain.Collection, len(dtos))

	for i, v := range dtos {
		requests := m.cleanRequests(v.Nodes)

		v.Context.Id = ""
		ctx := dto.ToContext(&v.Context)

		v.Id = ""
		v.Nodes = make([]dto.DtoNode, 0)
		collection := dto.ToCollection(&v)

		collection, err := m.insertResources(owner, collection, ctx, requests)
		if err != nil {
			return make([]domain.Collection, 0), err
		}

		collections[i] = *collection
	}

	return collections, nil
}

func (m *ManagerCollection) ResolveRequests(owner string, collection *domain.Collection, requests ...domain.Request) *domain.Collection {
	if len(requests) == 0 {
		return collection
	}

	m.mu.Lock()

	len := len(collection.Nodes)

	for i, v := range requests {
		node := domain.NodeReference{
			Order:   len + i,
			Request: v.Id,
		}

		collection.ResolveRequest(&node)
	}

	m.mu.Unlock()

	return m.Insert(owner, collection)
}

func (m *ManagerCollection) ImportDtoRequestsById(owner string, id string, dtos []dto.DtoRequest) *domain.Collection {
	collection, exists := m.collection.Find(id)
	if !exists || collection.Owner != owner {
		return nil
	}

	return m.ImportDtoRequests(owner, collection, dtos)
}

func (m *ManagerCollection) ImportDtoRequests(owner string, collection *domain.Collection, dtos []dto.DtoRequest) *domain.Collection {
	m.mu.Lock()

	if len(dtos) == 0 {
		return collection
	}

	len := len(collection.Nodes)

	requestStatus := domain.StatusCollectionToStatusRequest(&collection.Status)
	for i, v := range dtos {
		v.Id = ""
		v.Status = *requestStatus
		request := dto.ToRequest(&v)

		request = m.managerRequest.InsertRequest(owner, request)

		collection.Nodes = append(collection.Nodes, domain.NodeReference{
			Order:   len + i,
			Request: request.Id,
		})
	}

	m.mu.Unlock()

	return m.Insert(owner, collection)
}

func (m *ManagerCollection) ImportOpenApi(owner string, file []byte) (*domain.Collection, error) {
	oapi, raw, err := openapi.MakeFromJson(file)
	if err != nil {
		oapi, raw, err = openapi.MakeFromYaml(file)
		if err != nil {
			err = errors.New("the file provide has not valid format, it must be an JSON or YAML")
			return nil, err
		}
	}

	collection, ctx, requests, err := openapi.NewFactoryCollection(owner, oapi).SetRaw(*raw).Make()
	if err != nil {
		return nil, err
	}

	return m.insertResources(owner, collection, ctx, requests)
}

func (m *ManagerCollection) insertResources(owner string, collection *domain.Collection, ctx *context.Context, requests []domain.Request) (*domain.Collection, error) {
	m.mu.Lock()

	collection = m.collection.Insert(owner, collection)

	ctx = m.managerContext.Insert(owner, collection.Id, ctx)
	collection.Context = ctx.Id

	collection = m.collection.Insert(owner, collection)

	requestStatus := domain.StatusCollectionToStatusRequest(&collection.Status)
	for i, _ := range requests {
		requests[i].Status = *requestStatus
	}

	requests = m.managerRequest.InsertManyRequest(owner, requests)
	for i, v := range requests {
		node := domain.NodeReference{
			Order:   i,
			Request: v.Id,
		}
		collection.Nodes = append(collection.Nodes, node)
	}

	m.mu.Unlock()

	return m.Insert(owner, collection), nil
}

func (m *ManagerCollection) MoveRequestBetweenCollectionsById(owner, sourceId, targetId, requestId string, movement Movement) (*domain.Collection, *domain.Collection, *domain.Request) {
	source, exists := m.collection.Find(sourceId)
	if !exists {
		return nil, nil, nil
	}

	target, exists := m.collection.Find(targetId)
	if !exists {
		return nil, nil, nil
	}

	return m.MoveRequestBetweenCollections(owner, source, target, requestId, movement)
}

func (m *ManagerCollection) MoveRequestBetweenCollections(owner string, source, target *domain.Collection, requestId string, movement Movement) (*domain.Collection, *domain.Collection, *domain.Request) {
	if movement == MOVE {
		_, exists := source.TakeRequest(requestId)
		if exists {
			m.Insert(owner, source)
		}
	}

	request, exists := m.managerRequest.FindRequest(owner, requestId)
	if !exists {
		return nil, nil, nil
	}

	requestStatus := domain.StatusCollectionToStatusRequest(&target.Status)
	if request.Status != *requestStatus {
		request.Status = *requestStatus
		request = m.managerRequest.InsertRequest(owner, request)
	}

	target, request = m.collection.PushToCollection(owner, target, request)
	return source, target, request
}

func (m *ManagerCollection) CollectRequest(owner string, payload PayloadCollectRequest) (*domain.Collection, *domain.Request) {
	request := &payload.Request

	if source, exists := m.collection.Find(payload.SourceId); exists {
		_, exists := source.TakeRequest(request.Id)
		if exists {
			m.Insert(owner, source)
		}
	}

	target, exists := m.collection.Find(payload.TargetId)
	if !exists {
		target = domain.NewFreeCollection(owner)
		target.Name = payload.TargetName
		target = m.Insert(owner, target)
	}

	if payload.Movement == MOVE && request.Status != domain.DRAFT {
		request, _ = m.managerRequest.Delete(owner, request)
	}

	request.Id = ""

	request.Name = payload.RequestName
	request.Status = *domain.StatusCollectionToStatusRequest(&target.Status)
	request = m.managerRequest.InsertRequest(owner, request)

	return m.collection.PushToCollection(owner, target, request)
}

func (m *ManagerCollection) SortCollectionRequest(owner string, collection *domain.Collection, payload PayloadSortCollection) *domain.Collection {
	nodes := make([]domain.NodeReference, 0)
	for i, v := range payload.SortRequests().Nodes {
		node, exists := collection.TakeRequest(v.Request)
		if exists {
			node.Order = i
			nodes = append(nodes, *node)
		}
	}

	len := len(nodes)
	for i, v := range collection.Nodes {
		v.Order = i + len
		nodes = append(nodes, v)
	}

	collection.Nodes = nodes
	collection.SortRequests().FixRequestsOrder()

	collection = m.collection.Insert(owner, collection)

	return collection
}

func (m *ManagerCollection) RemoveRequestFromCollection(owner string, collection *domain.Collection, requestId string) (*domain.Collection, *domain.Request, *domain.Response) {
	_, exists := collection.TakeRequest(requestId)
	if exists {
		collection = m.Insert(owner, collection)
	}
	
	request, exists := m.managerRequest.FindRequest(owner, requestId)
	if !exists {
		return collection, nil, nil
	}

	request, response := m.managerRequest.Delete(owner, request)

	return collection, request, response
}

func (m *ManagerCollection) RemoveRequestFromCollectionById(owner string, collectionId string, requestId string) (*domain.Collection, *domain.Request, *domain.Response) {
	collection, exists := m.collection.Find(collectionId)
	if !exists || collection.Owner != owner {
		return nil, nil, nil
	}
	return m.RemoveRequestFromCollection(owner, collection, requestId)
}

func (m *ManagerCollection) CloneCollection(owner, id, name string) *domain.Collection {
	m.mu.Lock()

	collection, exists := m.collection.Find(id)
	if !exists || collection.Owner != owner {
		return nil
	}

	requestStatus := *domain.StatusCollectionToStatusRequest(&collection.Status)

	nodes := make([]domain.NodeReference, 0)
	for i, v := range collection.Nodes {
		request, exists := m.managerRequest.FindRequest(owner, v.Request)
		if !exists {
			continue
		}

		request.Id = ""
		request.Status = requestStatus
		request = m.managerRequest.InsertRequest(owner, request)
		
		nodes = append(nodes, domain.NodeReference{
			Order:   i,
			Request: request.Id,
		})
	}

	if context, exists := m.managerContext.Find(owner, collection.Context); exists {
		context.Id = ""
		context = m.managerContext.Insert(owner, collection.Id, context)
		collection.Context = context.Id
	} else {
		collection.Context = ""
	}

	collection.Id = ""
	collection.Name = name
	collection.Nodes = nodes
	collection.Timestamp = 0

	m.mu.Unlock()

	return m.Insert(owner, collection)
}

func (m *ManagerCollection) Delete(owner, id string) *domain.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, exists := m.collection.Find(id)
	if !exists || collection.Owner != owner {
		return nil
	}

	if context, exists := m.managerContext.Find(owner, collection.Context); !exists {
		m.managerContext.Delete(context)
	}

	requests := make([]string, len(collection.Nodes))
	for i, v := range collection.Nodes {
		requests[i] = v.Request
	}

	m.managerRequest.DeleteMany(owner, requests...)

	return m.collection.Delete(collection)
}

func (m *ManagerCollection) cleanRequests(dtos []dto.DtoNode) []domain.Request {
	requests := make([]domain.Request, len(dtos))

	for i, v := range dtos {
		v.Request.Id = ""
		requests[i] = *dto.ToRequest(&v.Request)
	}

	return requests
}
