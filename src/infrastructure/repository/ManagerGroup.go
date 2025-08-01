package repository

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type ManagerGroup struct {
	mu                sync.Mutex
	group             IRepositoryGroup
	managerCollection *ManagerCollection
}

func NewManagerGroup(group IRepositoryGroup, managerCollection *ManagerCollection) *ManagerGroup {
	return &ManagerGroup{
		group:             group,
		managerCollection: managerCollection,
	}
}

func (m *ManagerGroup) Find(owner, id string) (*domain.Group, bool) {
	return m.group.Find(id)
}

func (m *ManagerGroup) FindLiteNodes(owner string, group *domain.Group) []dto.DtoLiteNodeCollection {
	if group.Owner != owner {
		return make([]dto.DtoLiteNodeCollection, 0)
	}

	dtos := m.managerCollection.FindLiteCollectionNodes(owner, group.Nodes)

	if len(dtos) == len(group.Nodes) {
		return dtos
	}

	nodes := make([]domain.NodeReference, len(dtos))
	for _, v := range dtos {
		nodes = append(nodes, domain.NodeReference{
			Order: v.Order,
			Item:  v.Collection.Id,
		})
	}

	group.Nodes = nodes

	m.Insert(owner, group)

	return dtos
}

func (m *ManagerGroup) Insert(owner string, group *domain.Group) *domain.Group {
	if group.Owner != owner {
		return nil
	}

	group = group.SortNodes().FixNodesOrder()
	return m.group.Insert(owner, group)
}

func (m *ManagerGroup) ImportOpenApi(owner string, group *domain.Group, file []byte) (*domain.Group, *domain.Collection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, err := m.managerCollection.ImportOpenApi(owner, file)
	if err != nil {
		return nil, nil, err
	}

	if collection == nil {
		return group, nil, err
	}

	return m.ResolveCollectionReferences(owner, group, *collection), collection, nil
}

func (m *ManagerGroup) ImportDtoCollections(owner string, group *domain.Group, dtos ...dto.DtoCollection) (*domain.Group, []domain.Collection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collections, err := m.managerCollection.ImportDtoCollections(owner, dtos...)
	if err != nil {
		return nil, collections, err
	}

	return m.ResolveCollectionReferences(owner, group, collections...), collections, nil
}

func (m *ManagerGroup) ImportCollection(owner string, group *domain.Group, collection *domain.Collection) (*domain.Group, *domain.Collection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection = m.managerCollection.Insert(owner, collection)
	if collection == nil {
		return group, nil
	}

	return m.ResolveCollectionReferences(owner, group, *collection), collection
}

func (m *ManagerGroup) ImportDtoRequestsById(owner string, group *domain.Group, id string, dtos []dto.DtoRequest) (*domain.Group, *domain.Collection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection := m.managerCollection.ImportDtoRequestsById(owner, id, dtos...)
	if collection == nil {
		return group, nil
	}

	return m.ResolveCollectionReferences(owner, group, *collection), collection
}

func (m *ManagerGroup) CollectRequest(owner string, group *domain.Group, payload PayloadCollectRequest) (*domain.Group, *domain.Collection, *domain.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, request := m.managerCollection.CollectRequest(owner, payload)
	if collection == nil {
		return group, nil, nil
	}

	return m.ResolveCollectionReferences(owner, group, *collection), collection, request
}

func (m *ManagerGroup) ResolveCollectionReferences(owner string, group *domain.Group, collections ...domain.Collection) *domain.Group {
	if group.Owner != owner {
		return nil
	}

	if len(collections) == 0 {
		return group
	}
	
	for _, v := range collections {
		if v.Owner != owner {
			continue
		}

		group.ResolveNode(v.Id)
	}

	return m.Insert(owner, group)
}

func (m *ManagerGroup) CloneCollection(owner string, group *domain.Group, id, name string) (*domain.Group, *domain.Collection) {
	if group.Owner != owner {
		return nil, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	collection := m.managerCollection.CloneCollection(owner, id, name)
	if collection == nil {
		return nil, nil
	}

	return m.ResolveCollectionReferences(owner, group, *collection), collection
}

func (m *ManagerGroup) SortCollections(owner string, group *domain.Group, payload PayloadSortNodes) *domain.Group {
	if group.Owner != owner {
		return nil
	}

	nodes := make([]domain.NodeReference, 0)
	for i, v := range payload.SortNodes().Nodes {
		node, exists := group.TakeNode(v.Item)
		if exists {
			node.Order = i
			nodes = append(nodes, *node)
		}
	}

	len := len(nodes)
	for i, v := range group.Nodes {
		v.Order = i + len
		nodes = append(nodes, v)
	}

	group.Nodes = nodes
	group.SortNodes().FixNodesOrder()

	group = m.group.Insert(owner, group)

	return group
}

func (m *ManagerGroup) Delete(owner string, group *domain.Group) *domain.Group {
	if group.Owner != owner {
		return nil
	}

	return m.group.Delete(group)
}

func (m *ManagerGroup) DeleteCollection(owner string, group *domain.Group, id string) (*domain.Group, *domain.Collection) {
	if group.Owner != owner {
		return nil, nil
	}

	_, exists := group.TakeNode(id)
	if exists {
		group = m.Insert(owner, group)
	}

	collection, exists := m.managerCollection.Find(owner, id)
	if !exists || collection.Owner != owner {
		return group, nil
	}

	collection = m.managerCollection.Delete(owner, collection.Id)

	return group, collection
}
