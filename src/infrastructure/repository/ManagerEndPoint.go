package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
	"github.com/Rafael24595/go-collections/collection"
)

type ManagerEndPoint struct {
	endPoint IRepositoryEndPoint
}

func NewManagerEndPoint(endPoint IRepositoryEndPoint) *ManagerEndPoint {
	return &ManagerEndPoint{
		endPoint: endPoint,
	}
}

func (m *ManagerEndPoint) FindAll(owner string) []mock_domain.EndPointLite {
	endPoints := m.endPoint.FindAll(owner)
	return collection.VectorFromList(endPoints).
		Filter(func(e mock_domain.EndPointLite) bool {
			return e.Owner == owner
		}).
		Collect()
}

func (m *ManagerEndPoint) Find(owner, id string) (*mock_domain.EndPointFull, bool) {
	endPoint, ok := m.endPoint.Find(id)
	if !ok || endPoint.Owner != owner {
		return nil, false
	}

	full, _ := mock_domain.FullFromEndPoint(endPoint)
	return full, true
}

func (m *ManagerEndPoint) FindByRequest(owner string, method domain.HttpMethod, path string) (*mock_domain.EndPoint, bool) {
	endPoint, ok := m.endPoint.FindByRequest(owner, method, path)
	if !ok || endPoint.Owner != owner {
		return nil, false
	}

	return endPoint, true
}

func (m *ManagerEndPoint) Insert(owner string, endPoint *mock_domain.EndPointFull) (*mock_domain.EndPoint, []error) {
	if endPoint.Owner != owner {
		return nil, make([]error, 0)
	}

	full, errs := mock_domain.ToEndPointFromFull(endPoint)
	if len(errs) > 0 {
		return nil, errs
	}

	return m.endPoint.Insert(owner, full), make([]error, 0)
}

func (m *ManagerEndPoint) Delete(owner string, endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	if endPoint.Owner != owner {
		return nil
	}

	return m.endPoint.Delete(endPoint)
}
