package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain/mock"
)

type IRepositoryMetrics interface {
	Find(endPoint *mock.EndPoint) (*mock.Metrics, bool)
	Resolve(endPoint *mock.EndPoint, metrics *mock.Metrics) *mock.Metrics
	Delete(endPoint *mock.EndPoint, metrics *mock.Metrics) *mock.Metrics
}
