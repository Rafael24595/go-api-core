package repository

import (
	"strings"

	"github.com/Rafael24595/go-api-core/src/domain"
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-collections/collection"
)

type ManagerEndPoint struct {
	endPoint       IRepositoryEndPoint
	managerMetrics *ManagerMetrics
}

func NewManagerEndPoint(endPoint IRepositoryEndPoint, managerMetrics *ManagerMetrics) *ManagerEndPoint {
	return &ManagerEndPoint{
		endPoint:       endPoint,
		managerMetrics: managerMetrics,
	}
}

func (m *ManagerEndPoint) FindAll(owner string) []mock_domain.EndPointLite {
	endPoints := m.endPoint.FindAllLite(owner)
	return collection.VectorFromList(endPoints).
		Filter(func(e mock_domain.EndPointLite) bool {
			return e.Owner == owner
		}).
		Collect()
}

func (m *ManagerEndPoint) Find(owner, id string) (*mock_domain.EndPoint, bool) {
	endPoint, ok := m.endPoint.Find(id)
	if !ok || endPoint.Owner != owner {
		return nil, false
	}

	endPoint.Responses = mock_domain.FixResponses(endPoint.Responses)

	return endPoint, true
}

func (m *ManagerEndPoint) FindFull(owner, id string) (*mock_domain.EndPointFull, bool) {
	endPoint, ok := m.endPoint.Find(id)
	if !ok || endPoint.Owner != owner {
		return nil, false
	}

	endPoint.Responses = mock_domain.FixResponses(endPoint.Responses)

	full, _ := mock_domain.FullFromEndPoint(endPoint)
	return full, true
}

func (m *ManagerEndPoint) FindByRequest(owner string, method domain.HttpMethod, path string) (*mock_domain.EndPoint, bool) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	endPoint, ok := m.endPoint.FindByRequest(owner, method, path)
	if !ok || endPoint.Owner != owner {
		return nil, false
	}

	endPoint.Responses = mock_domain.FixResponses(endPoint.Responses)

	return endPoint, true
}

func (m *ManagerEndPoint) Insert(owner string, endPoint *mock_domain.EndPointFull) (*mock_domain.EndPoint, []error) {
	if endPoint.Owner != owner {
		return nil, make([]error, 0)
	}

	result, errs := mock_domain.ToEndPointFromFull(endPoint)
	if len(errs) > 0 {
		return nil, errs
	}

	result = mock_domain.FixEndPoint(owner, result)
	result.Responses = mock_domain.FixResponses(result.Responses)

	if !strings.HasPrefix(endPoint.Path, "/") {
		result.Path = "/" + endPoint.Path
	}

	return m.endPoint.Insert(result), make([]error, 0)
}

func (m *ManagerEndPoint) Delete(owner string, endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	if endPoint.Owner != owner {
		return nil
	}

	result := m.endPoint.Delete(endPoint)

	go m.managerMetrics.Delete(owner, endPoint)

	return result
}

func (m *ManagerEndPoint) Sort(owner string, references []domain.NodeReference) []mock_domain.EndPoint {
	endPoints := collection.VectorFromList(m.endPoint.FindAll(owner))

	sorted := make([]mock_domain.EndPoint, 0)
	for i, v := range references {
		endPoint, exists := endPoints.FindOne(func(e mock_domain.EndPoint) bool {
			return e.Id == v.Item
		})

		if !exists || endPoint.Owner != owner {
			continue
		}

		endPoint.Order = i
		sorted = append(sorted, *endPoint)
	}

	sorted = mock_domain.FixEndPoints(owner, sorted)

	return m.endPoint.InsertMany(sorted...)
}
