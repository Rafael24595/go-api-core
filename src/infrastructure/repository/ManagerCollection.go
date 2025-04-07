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

func (m *ManagerCollection) ImportDtoCollections(owner string, dtos []dto.DtoCollection) ([]domain.Collection, error) {
	collections := make([]domain.Collection, len(dtos))

	for i, v := range dtos {
		requests := m.cleanRequests(v.Nodes)
		
		v.Context.Id = "";
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

func (m *ManagerCollection) cleanRequests(dtos []dto.DtoNode) []domain.Request {
	requests := make([]domain.Request, len(dtos))

	for i, v := range dtos {
		v.Request.Id = ""
		requests[i] = *dto.ToRequest(&v.Request)
	}

	return requests
}

func (m *ManagerCollection) InsertOpenApi(owner string, file []byte) (*domain.Collection, error) {
	oapi, raw, err := openapi.MakeFromJson(file)
	if err != nil {
		oapi, raw, err = openapi.MakeFromYaml(file)
		if err != nil {
			err = errors.New("the file provide has not valid format, it must be an JSON or YAML")
			return nil, err
		}
	}

	collection, ctx, requests, err := openapi.NewFactoryCollection(owner, oapi).SetRaw(*raw).Make();
	if err != nil {
		return nil, err
	}

	return m.insertResources(owner, collection, ctx, requests)
}

func (m *ManagerCollection) insertResources(owner string, collection *domain.Collection, ctx *context.Context, requests []domain.Request) (*domain.Collection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection = m.collection.Insert(owner, collection)

	ctx = m.context.InsertFromCollection(owner, collection.Id, ctx)
	collection.Context = ctx.Id

	collection = m.collection.Insert(owner, collection)

	requests = m.request.InsertMany(owner, requests)
	for i, v := range requests {
		node := domain.NodeReference{
			Order: i,
			Request: v.Id,
		}
		collection.Nodes = append(collection.Nodes, node)
	}

	collection = m.collection.Insert(owner, collection)

	return collection, nil
}

func (m *ManagerCollection) Insert(owner string, collection *domain.Collection) *domain.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	if collection.Id == "" {
		collection = m.collection.Insert(owner, collection)
	}

	if _, exists := m.context.FindByCollection(owner, collection.Id); !exists {
		context := m.context.InsertFromCollection(owner, collection.Id, context.NewContext(owner))
		collection.Context = context.Id
	}

	return m.collection.Insert(owner, collection)
}

func (m *ManagerCollection) PushToCollection(owner string, payload PayloadPushToCollection) *domain.Collection {
	request := &payload.Request

	if source, exists := m.collection.Find(payload.SourceId); exists {
		_, exists := source.TakeRequest(request.Id)
		if exists {
			m.Insert(owner, source)
		}
	}

	if payload.Movement == MOVE && request.Status != domain.DRAFT  {
		request = m.request.Delete(request)
	}

	request.Id = ""
	
	request.Name = payload.RequestName
	request.Status = domain.GROUP
	request = m.request.Insert(owner, request)

	collection, exists := m.collection.Find(payload.TargetId)
	if !exists {
		collection = domain.NewCollection(owner)
		collection.Name = payload.TargetName
		collection = m.Insert(owner, collection)
	}

	return m.collection.PushToCollection(owner, collection, request)
}

func (m *ManagerCollection) TakeFromCollection(owner, collectionId, requestId string) (*domain.Collection, *domain.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, exists := m.collection.Find(collectionId)
	if !exists || collection.Owner != owner {
		return nil, nil
	}

	_, exists = collection.TakeRequest(requestId)
	if !exists {
		return nil, nil
	}

	collection = m.collection.Insert(owner, collection)

	request, exists := m.request.Find(requestId)
	if !exists {
		return collection, nil
	}

	request.Status = domain.FINAL
	request = m.request.Insert(owner, request)

	return collection, request
}

func (m *ManagerCollection) RemoveFromCollection(owner string, collectionId string, requestId string) (*domain.Collection, *domain.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, exists := m.collection.Find(collectionId)
	if !exists || collection.Owner != owner {
		return nil, nil
	}

	_, exists = collection.TakeRequest(requestId)
	if !exists {
		return nil, nil
	}

	collection = m.collection.Insert(owner, collection)

	request, exists := m.request.Find(requestId)
	if !exists {
		return collection, nil
	}

	request = m.request.Delete(request)

	return collection, request
}

func (m *ManagerCollection) CloneCollection(owner, id, name string) *domain.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, exists := m.collection.Find(id)
	if !exists || collection.Owner != owner {
		return nil
	}

	nodes := make([]domain.NodeReference, 0)
	for i, v := range collection.Nodes {
		request, exists := m.request.Find(v.Request)
		if !exists {
			continue
		}
		request.Id = ""
		request = m.request.Insert(owner, request)
		nodes = append(nodes, domain.NodeReference{
			Order: i,
			Request: request.Id,
		})
	}
	
	if context, exists := m.context.Find(collection.Context); exists {
		context.Id = ""
		context = m.context.InsertFromCollection(owner, collection.Id, context)
		collection.Context = context.Id
	} else {
		collection.Context = ""
	}

	collection.Id = ""
	collection.Name = name
	collection.Nodes = nodes
	collection.Timestamp = 0

	return m.collection.Insert(owner, collection)
}

func (m *ManagerCollection) Delete(owner, id string) *domain.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, exists := m.collection.Find(id)
	if !exists || collection.Owner != owner {
		return nil
	}

	if context, exists := m.context.Find(collection.Context); !exists {
		m.context.Delete(context)
	}

	requests := make([]string, len(collection.Nodes))
	for i, v := range collection.Nodes {
		requests[i] = v.Request
	}

	m.request.DeleteMany(requests...)
	m.response.DeleteMany(requests...)

	return m.collection.Delete(collection)
}
