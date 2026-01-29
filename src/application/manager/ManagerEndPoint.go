package manager

import (
	"strings"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-collections/collection"
)

type ManagerEndPoint struct {
	endPoint       mock.RepositoryEndPoint
	managerMetrics *ManagerMetrics
}

func NewManagerEndPoint(endPoint mock.RepositoryEndPoint, managerMetrics *ManagerMetrics) *ManagerEndPoint {
	return &ManagerEndPoint{
		endPoint:       endPoint,
		managerMetrics: managerMetrics,
	}
}

func (m *ManagerEndPoint) Export(owner string) []mock.EndPoint {
	return m.endPoint.FindAll(owner)
}

func (m *ManagerEndPoint) ExportList(owner string, ids ...string) []mock.EndPoint {
	endPoints := m.endPoint.FindMany(ids...)
	return collection.VectorFromList(endPoints).
		Filter(func(e mock.EndPoint) bool {
			return e.Owner == owner
		}).
		Collect()
}

func (m *ManagerEndPoint) Import(owner string, endPoints []mock.EndPoint) []string {
	vec := collection.VectorFromList(endPoints).
		Map(func(i int, e mock.EndPoint) mock.EndPoint {
			return *mock.CleanEndPoint(owner, &e)
		})

	result := m.endPoint.InsertMany(vec.Collect()...)

	return collection.MapToVector(result, func(e mock.EndPoint) string {
		return e.Id
	}).Collect()
}

func (m *ManagerEndPoint) FindAll(owner string) []mock.EndPointLite {
	endPoints := m.endPoint.FindAll(owner)
	return collection.MapToVector(endPoints, func(e mock.EndPoint) mock.EndPointLite {
		return *mock.LiteFromEndPoint(&e)
	}).Collect()
}

func (m *ManagerEndPoint) Find(owner, id string) (*mock.EndPoint, bool) {
	endPoint, ok := m.endPoint.Find(id)
	if !ok || endPoint.Owner != owner {
		return nil, false
	}

	endPoint.Responses = mock.FixResponses(endPoint.Responses)

	return endPoint, true
}

func (m *ManagerEndPoint) FindFull(owner, id string) (*mock.EndPointFull, bool) {
	endPoint, ok := m.endPoint.Find(id)
	if !ok || endPoint.Owner != owner {
		return nil, false
	}

	endPoint.Responses = mock.FixResponses(endPoint.Responses)

	full, _ := mock.FullFromEndPoint(endPoint)
	return full, true
}

func (m *ManagerEndPoint) FindByRequest(owner string, method domain.HttpMethod, path string) (*mock.EndPoint, bool) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	endPoint, ok := m.endPoint.FindByRequest(owner, method, path)
	if !ok || endPoint.Owner != owner {
		return nil, false
	}

	endPoint.Responses = mock.FixResponses(endPoint.Responses)

	return endPoint, true
}

func (m *ManagerEndPoint) Insert(owner string, endPoint *mock.EndPointFull) (*mock.EndPoint, []error) {
	if endPoint.Owner != owner {
		return nil, make([]error, 0)
	}

	result, errs := mock.ToEndPointFromFull(endPoint)
	if len(errs) > 0 {
		return nil, errs
	}

	result = mock.FixEndPoint(owner, result)
	result.Responses = mock.FixResponses(result.Responses)

	if !strings.HasPrefix(endPoint.Path, "/") {
		result.Path = "/" + endPoint.Path
	}

	return m.endPoint.Insert(result), make([]error, 0)
}

func (m *ManagerEndPoint) Delete(owner string, id string) *mock.EndPoint {
	endPoint, ok := m.endPoint.Find(id)
	if !ok || endPoint.Owner != owner {
		return nil
	}

	result := m.endPoint.Delete(endPoint)

	go m.managerMetrics.Delete(owner, endPoint)

	return result
}

func (m *ManagerEndPoint) Sort(owner string, references []domain.NodeReference) []mock.EndPoint {
	endPoints := collection.VectorFromList(m.endPoint.FindAll(owner))

	sorted := make([]mock.EndPoint, 0)
	for i, v := range references {
		endPoint, exists := endPoints.FindOne(func(e mock.EndPoint) bool {
			return e.Id == v.Item
		})

		if !exists || endPoint.Owner != owner {
			continue
		}

		endPoint.Order = i
		sorted = append(sorted, endPoint)
	}

	sorted = mock.FixEndPoints(owner, sorted)

	return m.endPoint.InsertMany(sorted...)
}
