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

func (m *ManagerCollection) Find(owner string, id string) (*domain.Collection, bool) {
	collection, exists := m.collection.Find(id)
	if !exists || collection.Owner != owner {
		return nil, false
	}
	return collection, exists
}

func (m *ManagerCollection) FindRequestNodes(owner string, collection *domain.Collection) []dto.DtoNodeRequest {
	if collection.Owner != owner {
		return make([]dto.DtoNodeRequest, 0)
	}

	dtos := m.managerRequest.FindNodes(owner, collection.Nodes)
	if len(dtos) == len(collection.Nodes) {
		return dtos
	}

	nodes := make([]domain.NodeReference, 0)
	for _, v := range dtos {
		nodes = append(nodes, domain.NodeReference{
			Order: v.Order,
			Item:  v.Request.Id,
		})
	}

	collection.Nodes = nodes

	m.Insert(owner, collection)

	return dtos
}

func (m *ManagerCollection) FindCollectionNodes(owner string, nodes []domain.NodeReference) []dto.DtoNodeCollection {
	collections := m.collection.FindCollections(nodes)
	
	dtos := make([]dto.DtoNodeCollection, 0)
	for _, v := range collections {
		collection := v.Collection
		if collection.Owner != owner {
			continue
		}

		requests := m.managerRequest.FindNodes(owner, collection.Nodes)
		context, _ := m.managerContext.Find(owner, collection.Context)
		if context == nil {
			continue
		}
		
		dtoContext := dto.FromContext(context)
		dtos = append(dtos, dto.DtoNodeCollection{
			Order:   v.Order,
			Collection: *dto.FromCollection(&collection, dtoContext, requests),
		})
	}
	
	return dtos
}

func (m *ManagerCollection) Insert(owner string, collection *domain.Collection) *domain.Collection {
	if m.isNotOwner(owner, collection) {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if collection.Id == "" {
		collection = m.collection.Insert(owner, collection)
	}

	if _, exists := m.managerContext.Find(owner, collection.Context); !exists {
		context := m.managerContext.Insert(owner, collection, context.NewContext(owner))
		collection.Context = context.Id
	}

	collection = collection.SortRequests().FixRequestsOrder()

	return m.collection.Insert(owner, collection)
}

func (m *ManagerCollection) CollectRequest(owner string, payload PayloadCollectRequest) (*domain.Collection, *domain.Request) {
	request := &payload.Request
	if request.Owner != owner {
		return nil, nil
	}

	target, exists := m.collection.Find(payload.TargetId)
	if !exists {
		target = domain.NewFreeCollection(owner)
		target.Name = payload.TargetName
		target = m.Insert(owner, target)
	}

	if exists && target.Owner != owner {
		return nil, nil
	}

	if payload.Movement == MOVE && request.Status != domain.DRAFT {
		result, isOwner := m.moveRequest(owner, payload, request)
		if !isOwner {
			return nil, nil
		}
		request = result
	}

	request.Id = ""

	request.Name = payload.RequestName
	request.Status = *domain.StatusCollectionToStatusRequest(&target.Status)
	request = m.managerRequest.InsertRequest(owner, request)

	return m.collection.PushToCollection(owner, target, request)
}

func (m *ManagerCollection) moveRequest(owner string, payload PayloadCollectRequest, request *domain.Request) (*domain.Request, bool) {
	if source, exists := m.collection.Find(payload.SourceId); exists {
		if source.Owner != owner {
			return request, false
		}

		_, exists := source.TakeRequest(request.Id)
		if exists {
			m.Insert(owner, source)
		}
	}

	request, _ = m.managerRequest.Delete(owner, request)

	return request, true
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

func (m *ManagerCollection) ImportDtoCollections(owner string, dtos ...dto.DtoCollection) ([]domain.Collection, error) {
	collections := make([]domain.Collection, len(dtos))

	for i, v := range dtos {
		requests := m.dtoNodeRequestToRequest(v.Nodes)

		v.Context.Id = ""
		ctx := dto.ToContext(&v.Context)

		v.Id = ""
		v.Nodes = make([]dto.DtoNodeRequest, 0)
		collection := dto.ToCollection(&v)

		collection, err := m.insertResources(owner, collection, ctx, requests)
		if err != nil {
			return make([]domain.Collection, 0), err
		}

		collections[i] = *collection
	}

	return collections, nil
}

func (m *ManagerCollection) ImportDtoRequestsById(owner string, id string, dtos ...dto.DtoRequest) *domain.Collection {
	collection, exists := m.collection.Find(id)
	if !exists {
		return nil
	}
	return m.ImportDtoRequests(owner, collection, dtos)
}

func (m *ManagerCollection) ImportDtoRequests(owner string, collection *domain.Collection, dtos []dto.DtoRequest) *domain.Collection {
	if collection.Owner != owner {
		return nil
	}

	m.mu.Lock()

	if len(dtos) == 0 {
		return collection
	}

	requestStatus := domain.StatusCollectionToStatusRequest(&collection.Status)
	requests := make([]domain.Request, len(dtos))
	for i, v := range dtos {
		v.Id = ""
		v.Status = *requestStatus
		requests[i] = *dto.ToRequest(&v)
	}
	
	requests = m.managerRequest.InsertManyRequest(owner, requests)

	len := len(collection.Nodes)
	for i, v := range requests {
		collection.Nodes = append(collection.Nodes, domain.NodeReference{
			Order: len + i,
			Item:  v.Id,
		})
	}

	m.mu.Unlock()

	return m.Insert(owner, collection)
}

func (m *ManagerCollection) insertResources(owner string, collection *domain.Collection, ctx *context.Context, requests []domain.Request) (*domain.Collection, error) {
	m.mu.Lock()

	collection = m.collection.Insert(owner, collection)

	ctx = m.managerContext.Insert(owner, collection, ctx)
	collection.Context = ctx.Id

	collection = m.collection.Insert(owner, collection)

	requestStatus := domain.StatusCollectionToStatusRequest(&collection.Status)
	for i := range requests {
		requests[i].Status = *requestStatus
	}

	requests = m.managerRequest.InsertManyRequest(owner, requests)
	for i, v := range requests {
		node := domain.NodeReference{
			Order: i,
			Item:  v.Id,
		}
		collection.Nodes = append(collection.Nodes, node)
	}

	m.mu.Unlock()

	return m.Insert(owner, collection), nil
}

func (m *ManagerCollection) ResolveRequestReferences(owner string, collection *domain.Collection, requests ...domain.Request) *domain.Collection {
	if collection.Owner != owner {
		return nil
	}

	if len(requests) == 0 {
		return collection
	}

	m.mu.Lock()

	for _, v := range requests {
		if v.Owner != owner {
			continue
		}

		collection.ResolveRequest(v.Id)
	}

	m.mu.Unlock()

	return m.Insert(owner, collection)
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
	if source.Owner != owner || target.Owner != owner {
		return nil, nil, nil
	}

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

func (m *ManagerCollection) SortCollectionRequestById(owner string, id string, payload PayloadSortNodes) *domain.Collection {
	collection, exists := m.collection.Find(id)
	if !exists {
		return nil
	}
	return m.SortCollectionRequest(owner, collection, payload)
}

func (m *ManagerCollection) SortCollectionRequest(owner string, collection *domain.Collection, payload PayloadSortNodes) *domain.Collection {
	if collection.Owner != owner {
		return nil
	}

	nodes := make([]domain.NodeReference, 0)
	for i, v := range payload.SortNodes().Nodes {
		node, exists := collection.TakeRequest(v.Item)
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

func (m *ManagerCollection) CloneCollection(owner, id, name string) *domain.Collection {
	collection, exists := m.collection.Find(id)
	if !exists || collection.Owner != owner {
		return nil
	}

	m.mu.Lock()

	requestStatus := *domain.StatusCollectionToStatusRequest(&collection.Status)

	nodeRequests := m.managerRequest.FindRequests(owner, collection.Nodes)
	requests := make([]domain.Request, len(nodeRequests))
	for i, v := range nodeRequests {
		v.Id = ""
		v.Status = requestStatus
		requests[i] = v
	}

	requests = m.managerRequest.InsertManyRequest(owner, requests)
	nodes := make([]domain.NodeReference, len(requests))
	for i, v := range requests {
		nodes = append(nodes, domain.NodeReference{
			Order: i,
			Item:  v.Id,
		})
	}

	if context, exists := m.managerContext.Find(owner, collection.Context); exists {
		context.Id = ""
		context = m.managerContext.Insert(owner, collection, context)
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
		m.managerContext.Delete(owner, context)
	}

	requests := make([]string, len(collection.Nodes))
	for i, v := range collection.Nodes {
		requests[i] = v.Item
	}

	m.managerRequest.DeleteMany(owner, requests...)

	return m.collection.Delete(collection)
}

func (m *ManagerCollection) DeleteRequestFromCollectionById(owner string, collectionId string, requestId string) (*domain.Collection, *domain.Request, *domain.Response) {
	collection, exists := m.collection.Find(collectionId)
	if !exists {
		return nil, nil, nil
	}
	return m.DeleteRequestFromCollection(owner, collection, requestId)
}

func (m *ManagerCollection) DeleteRequestFromCollection(owner string, collection *domain.Collection, requestId string) (*domain.Collection, *domain.Request, *domain.Response) {
	if collection.Owner != owner {
		return nil, nil, nil
	}

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

func (m *ManagerCollection) dtoNodeRequestToRequest(dtos []dto.DtoNodeRequest) []domain.Request {
	requests := make([]domain.Request, len(dtos))

	for i, v := range dtos {
		v.Request.Id = ""
		requests[i] = *dto.ToRequest(&v.Request)
	}

	return requests
}

func (m *ManagerCollection) isNotOwner(owner string, collection *domain.Collection) bool {
	return !m.isOwner(owner, collection)
}

func (m *ManagerCollection) isOwner(owner string, collection *domain.Collection) bool {
	if (collection == nil) {
		return false
	}

	if (collection.Id != "" && collection.Owner != owner) {
		return false
	}
	
	return true
}
