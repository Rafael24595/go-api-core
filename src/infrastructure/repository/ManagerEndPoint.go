package repository

import (
	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
)

type ManagerEndPoint struct {
	endPoint IRepositoryEndPoint
}

func NewManagerEndPoint(endPoint IRepositoryEndPoint) *ManagerEndPoint {
	return &ManagerEndPoint{
		endPoint: endPoint,
	}
}

func (m *ManagerEndPoint) Find(owner, id string) (*mock_domain.EndPoint, bool) {
	endPoint, ok := m.endPoint.Find(id)
	if ok && endPoint.Owner != owner {
		return nil, false
	}
	return endPoint, ok
}

func (m *ManagerEndPoint) Insert(owner string, endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	if endPoint.Owner != owner {
		return nil
	}

	return m.endPoint.Insert(owner, endPoint)
}

func (m *ManagerEndPoint) Delete(owner string, endPoint *mock_domain.EndPoint) *mock_domain.EndPoint {
	if endPoint.Owner != owner {
		return nil
	}

	return m.endPoint.Delete(endPoint)
}
