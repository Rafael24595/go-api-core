package repository

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/collection"
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

func (m *ManagerCollection) Find(owner string, id string) (*collection.Collection, bool) {
	collection, exists := m.collection.Find(id)
	if !exists || collection.Owner != owner {
		return nil, false
	}
	return collection, exists
}

func (m *ManagerCollection) FindDto(owner string, id string) (*dto.DtoCollection, bool) {
	collection, exists := m.Find(owner, id)
	if !exists {
		return nil, false
	}

	requests := m.managerRequest.FindNodes(owner, collection.Nodes)
	context, _ := m.managerContext.Find(owner, collection.Context)
	if context == nil {
		return nil, false
	}

	dtoContext := dto.FromContext(context)

	dtoCollection := dto.FromCollection(collection, dtoContext, requests)

	return dtoCollection, dtoCollection != nil
}

func (m *ManagerCollection) FindDtoLite(owner string, id string) (*dto.DtoLiteCollection, bool) {
	collection, exists := m.Find(owner, id)
	if !exists {
		return nil, false
	}
	return m.makeLiteCollection(owner, *collection), false
}

func (m *ManagerCollection) FindLiteRequestNodes(owner string, collection *collection.Collection) []dto.DtoLiteNodeRequest {
	if collection.Owner != owner {
		return make([]dto.DtoLiteNodeRequest, 0)
	}

	dtos := m.managerRequest.FindLiteNodes(owner, collection.Nodes)
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

func (m *ManagerCollection) FindLiteCollectionNodes(owner string, nodes []domain.NodeReference) []dto.DtoLiteNodeCollection {
	collections := m.collection.FindCollections(nodes)

	dtos := make([]dto.DtoLiteNodeCollection, 0)
	for _, v := range collections {
		collection := m.makeLiteCollection(owner, v.Collection)
		if collection == nil {
			continue
		}

		dtos = append(dtos, dto.DtoLiteNodeCollection{
			Order:      v.Order,
			Collection: *collection,
		})
	}

	return dtos
}

func (m *ManagerCollection) makeLiteCollection(owner string, collection collection.Collection) *dto.DtoLiteCollection {
	if collection.Owner != owner {
		return nil
	}

	requests := m.managerRequest.FindLiteNodes(owner, collection.Nodes)
	context, _ := m.managerContext.Find(owner, collection.Context)
	if context == nil {
		return nil
	}

	dtoContext := dto.FromContext(context)

	return dto.ToLiteCollection(&collection, dtoContext, requests)
}

func (m *ManagerCollection) Insert(owner string, collection *collection.Collection) *collection.Collection {
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

func (m *ManagerCollection) CollectRequest(owner string, payload PayloadCollectRequest) (*collection.Collection, *action.Request) {
	request := &payload.Request
	if request.Owner != owner {
		return nil, nil
	}

	target, exists := m.collection.Find(payload.TargetId)
	if !exists {
		target = collection.NewFreeCollection(owner)
		target.Name = payload.TargetName
		target = m.Insert(owner, target)
	}

	if exists && target.Owner != owner {
		return nil, nil
	}

	if payload.Movement == MOVE && request.Status != action.DRAFT {
		result, isOwner := m.moveRequest(owner, payload, request)
		if !isOwner {
			return nil, nil
		}
		request = result
	}

	request.Id = ""

	request.Name = payload.RequestName
	request.Status = *collection.StatusCollectionToStatusRequest(&target.Status)
	request = m.managerRequest.InsertRequest(owner, request)

	return m.collection.PushToCollection(owner, target, request)
}

func (m *ManagerCollection) moveRequest(owner string, payload PayloadCollectRequest, request *action.Request) (*action.Request, bool) {
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

func (m *ManagerCollection) ImportOpenApi(owner string, file []byte) (*collection.Collection, error) {
	oapi, raw, err := openapi.MakeFromJson(file)
	if err != nil {
		oapi, raw, err = openapi.MakeFromYaml(file)
		if err != nil {
			err = errors.New("the provided file has an invalid format, it must be an JSON or YAML")
			return nil, err
		}
	}

	version, err := utils.ParseVersion(oapi.OpenAPI)
	if err != nil {
		return nil, err
	}

	if version.Major < 3 {
		//TODO: Add support to previous and future versions.
		err = fmt.Errorf("the provided file has an invalid version '%s'; it must be 3.0.0 or higher", oapi.Info.Version)
		return nil, err
	}

	collection, ctx, requests, err := openapi.NewFactoryCollection(owner, oapi).SetRaw(*raw).Make()
	if err != nil {
		return nil, err
	}

	return m.insertResources(owner, collection, ctx, requests)
}

func (m *ManagerCollection) ImportDtoCollections(owner string, dtos ...dto.DtoCollection) ([]collection.Collection, error) {
	collections := make([]collection.Collection, len(dtos))

	for i, v := range dtos {
		requests := m.dtoNodeRequestToRequest(v.Nodes)

		v.Context.Id = ""
		ctx := dto.ToContext(&v.Context)

		v.Id = ""
		v.Nodes = make([]dto.DtoNodeRequest, 0)
		coll := dto.ToCollection(&v)

		coll, err := m.insertResources(owner, coll, ctx, requests)
		if err != nil {
			return make([]collection.Collection, 0), err
		}

		collections[i] = *coll
	}

	return collections, nil
}

func (m *ManagerCollection) ImportDtoRequestsById(owner string, id string, dtos ...dto.DtoRequest) *collection.Collection {
	collection, exists := m.collection.Find(id)
	if !exists {
		return nil
	}
	return m.ImportDtoRequests(owner, collection, dtos)
}

func (m *ManagerCollection) ImportDtoRequests(owner string, coll *collection.Collection, dtos []dto.DtoRequest) *collection.Collection {
	if coll.Owner != owner {
		return nil
	}

	m.mu.Lock()

	if len(dtos) == 0 {
		return coll
	}

	requestStatus := collection.StatusCollectionToStatusRequest(&coll.Status)
	requests := make([]action.Request, len(dtos))
	for i, v := range dtos {
		v.Id = ""
		v.Status = *requestStatus
		requests[i] = *dto.ToRequest(&v)
	}

	requests = m.managerRequest.InsertManyRequest(owner, requests)

	len := len(coll.Nodes)
	for i, v := range requests {
		coll.Nodes = append(coll.Nodes, domain.NodeReference{
			Order: len + i,
			Item:  v.Id,
		})
	}

	m.mu.Unlock()

	return m.Insert(owner, coll)
}

func (m *ManagerCollection) insertResources(owner string, coll *collection.Collection, ctx *context.Context, requests []action.Request) (*collection.Collection, error) {
	m.mu.Lock()

	coll = m.collection.Insert(owner, coll)

	ctx = m.managerContext.Insert(owner, coll, ctx)
	coll.Context = ctx.Id

	coll = m.collection.Insert(owner, coll)

	requestStatus := collection.StatusCollectionToStatusRequest(&coll.Status)
	for i := range requests {
		requests[i].Status = *requestStatus
	}

	requests = m.managerRequest.InsertManyRequest(owner, requests)
	for i, v := range requests {
		node := domain.NodeReference{
			Order: i,
			Item:  v.Id,
		}
		coll.Nodes = append(coll.Nodes, node)
	}

	m.mu.Unlock()

	return m.Insert(owner, coll), nil
}

func (m *ManagerCollection) ResolveRequestReferences(owner string, collection *collection.Collection, requests ...action.Request) *collection.Collection {
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

func (m *ManagerCollection) MoveRequestBetweenCollectionsById(owner, sourceId, targetId, requestId string, movement Movement) (*collection.Collection, *collection.Collection, *action.Request) {
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

func (m *ManagerCollection) MoveRequestBetweenCollections(owner string, source, target *collection.Collection, requestId string, movement Movement) (*collection.Collection, *collection.Collection, *action.Request) {
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

	requestStatus := collection.StatusCollectionToStatusRequest(&target.Status)
	if request.Status != *requestStatus {
		request.Status = *requestStatus
		request = m.managerRequest.InsertRequest(owner, request)
	}

	target, request = m.collection.PushToCollection(owner, target, request)
	return source, target, request
}

func (m *ManagerCollection) SortCollectionRequestById(owner string, id string, payload PayloadSortNodes) *collection.Collection {
	collection, exists := m.collection.Find(id)
	if !exists {
		return nil
	}
	return m.SortCollectionRequest(owner, collection, payload)
}

func (m *ManagerCollection) SortCollectionRequest(owner string, collection *collection.Collection, payload PayloadSortNodes) *collection.Collection {
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

func (m *ManagerCollection) CloneCollection(owner, id, name string) *collection.Collection {
	coll, exists := m.collection.Find(id)
	if !exists || coll.Owner != owner {
		return nil
	}

	m.mu.Lock()

	requestStatus := *collection.StatusCollectionToStatusRequest(&coll.Status)

	nodeRequests := m.managerRequest.FindRequests(owner, coll.Nodes)
	requests := make([]action.Request, len(nodeRequests))
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

	if context, exists := m.managerContext.Find(owner, coll.Context); exists {
		context.Id = ""
		context = m.managerContext.Insert(owner, coll, context)
		coll.Context = context.Id
	} else {
		coll.Context = ""
	}

	coll.Id = ""
	coll.Name = name
	coll.Nodes = nodes
	coll.Timestamp = 0

	m.mu.Unlock()

	return m.Insert(owner, coll)
}

func (m *ManagerCollection) Delete(owner, id string) *collection.Collection {
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

func (m *ManagerCollection) DeleteRequestFromCollectionById(owner string, collectionId string, requestId string) (*collection.Collection, *action.Request, *action.Response) {
	collection, exists := m.collection.Find(collectionId)
	if !exists {
		return nil, nil, nil
	}
	return m.DeleteRequestFromCollection(owner, collection, requestId)
}

func (m *ManagerCollection) DeleteRequestFromCollection(owner string, collection *collection.Collection, requestId string) (*collection.Collection, *action.Request, *action.Response) {
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

func (m *ManagerCollection) dtoNodeRequestToRequest(dtos []dto.DtoNodeRequest) []action.Request {
	requests := make([]action.Request, len(dtos))

	for i, v := range dtos {
		v.Request.Id = ""
		requests[i] = *dto.ToRequest(&v.Request)
	}

	return requests
}

func (m *ManagerCollection) isNotOwner(owner string, collection *collection.Collection) bool {
	return !m.isOwner(owner, collection)
}

func (m *ManagerCollection) isOwner(owner string, collection *collection.Collection) bool {
	if collection == nil {
		return false
	}

	if collection.Id != "" && collection.Owner != owner {
		return false
	}

	return true
}
