package manager

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/domain/group"
	"github.com/Rafael24595/go-api-core/src/infrastructure/dto"
)

type ManagerGroup struct {
	mu                sync.Mutex
	group             group.Repository
	managerCollection *ManagerCollection
}

func NewManagerGroup(group group.Repository, managerCollection *ManagerCollection) *ManagerGroup {
	return &ManagerGroup{
		group:             group,
		managerCollection: managerCollection,
	}
}

func (m *ManagerGroup) Find(owner, id string) (*group.Group, bool) {
	return m.group.Find(id)
}

func (m *ManagerGroup) FindLiteNodes(owner string, group *group.Group) []collection.NodeCollectionLite {
	if group.Owner != owner {
		return make([]collection.NodeCollectionLite, 0)
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

func (m *ManagerGroup) Insert(owner string, group *group.Group) *group.Group {
	if group.Owner != owner {
		return nil
	}

	group = group.SortNodes().FixNodesOrder()
	return m.group.Insert(owner, group)
}

func (m *ManagerGroup) ImportOpenApi(owner string, group *group.Group, file []byte) (*group.Group, *collection.Collection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, err := m.managerCollection.ImportOpenApi(owner, file)
	if err != nil {
		return nil, nil, err
	}

	if collection == nil {
		return group, nil, err
	}

	return m.resolveCollectionReferences(owner, group, *collection), collection, nil
}

func (m *ManagerGroup) ImportDtoCollections(owner string, group *group.Group, dtos ...dto.DtoCollection) (*group.Group, []collection.Collection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collections, err := m.managerCollection.ImportDtoCollections(owner, dtos...)
	if err != nil {
		return nil, collections, err
	}

	return m.resolveCollectionReferences(owner, group, collections...), collections, nil
}

func (m *ManagerGroup) ImportCollection(owner string, group *group.Group, collection *collection.Collection) (*group.Group, *collection.Collection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection = m.managerCollection.Insert(owner, collection)
	if collection == nil {
		return group, nil
	}

	return m.resolveCollectionReferences(owner, group, *collection), collection
}

func (m *ManagerGroup) ImportRequestsById(owner string, group *group.Group, id string, reqs ...action.Request) (*group.Group, *collection.Collection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection := m.managerCollection.ImportRequestsById(owner, id, reqs...)
	if collection == nil {
		return group, nil
	}

	return m.resolveCollectionReferences(owner, group, *collection), collection
}

func (m *ManagerGroup) CollectRequest(owner string, group *group.Group, payload PayloadCollectRequest) (*group.Group, *collection.Collection, *action.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, request := m.managerCollection.CollectRequest(owner, payload)
	if collection == nil {
		return group, nil, nil
	}

	return m.resolveCollectionReferences(owner, group, *collection), collection, request
}

func (m *ManagerGroup) CloneCollection(owner string, group *group.Group, id, name string) (*group.Group, *collection.Collection) {
	if group.Owner != owner {
		return nil, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	collection := m.managerCollection.CloneCollection(owner, id, name)
	if collection == nil {
		return nil, nil
	}

	return m.resolveCollectionReferences(owner, group, *collection), collection
}

func (m *ManagerGroup) resolveCollectionReferences(owner string, group *group.Group, collections ...collection.Collection) *group.Group {
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

func (m *ManagerGroup) SortCollections(owner string, group *group.Group, payload PayloadSortNodes) *group.Group {
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

func (m *ManagerGroup) Delete(owner string, id string) *group.Group {
	group, exists := m.group.Find(id)
	if !exists || group.Owner != owner {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, v := range group.Nodes {
		m.managerCollection.Delete(owner, v.Item)
	}

	return m.group.Delete(group)
}

func (m *ManagerGroup) DeleteCollection(owner string, group *group.Group, id string) (*group.Group, *collection.Collection) {
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
