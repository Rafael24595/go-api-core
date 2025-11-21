package repository

import (
	"fmt"
	"sync"
	"time"

	mock_domain "github.com/Rafael24595/go-api-core/src/domain/mock"
)

type ManagerMetrics struct {
	mu      sync.Mutex
	metrics IRepositoryMetrics
}

func NewManagerMetrics(metrics IRepositoryMetrics) *ManagerMetrics {
	return &ManagerMetrics{
		metrics: metrics,
	}
}

func (m *ManagerMetrics) Find(owner string, endPoint *mock_domain.EndPoint) (*mock_domain.Metrics, bool) {
	if endPoint.Owner != owner {
		return nil, false
	}

	metrics, ok := m.metrics.Find(endPoint)
	if !ok {
		metrics = mock_domain.EmptyMetrics(endPoint)
	}

	if metrics.LastStarted == 0 && endPoint.Status {
		metrics.LastStarted = endPoint.Modified
		metrics.TotalUptime = endPoint.Modified
	}

	if metrics.LastStarted != 0 {
		metrics.TotalUptime += (time.Now().UnixMilli() - metrics.LastStarted)
	}

	return metrics, true
}

func (m *ManagerMetrics) ResolveStatus(owner string, oldEndPoint *mock_domain.EndPoint, newEndPoint *mock_domain.EndPoint) *mock_domain.Metrics {
	if oldEndPoint.Owner != owner || oldEndPoint.Owner != newEndPoint.Owner {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, ok := m.metrics.Find(newEndPoint)
	if !ok {
		metrics = mock_domain.EmptyMetrics(newEndPoint)
	}

	now := time.Now().UnixMilli()

	if !oldEndPoint.Status && newEndPoint.Status {
		metrics.LastStarted = now
	}

	if oldEndPoint.Status && !newEndPoint.Status {
		if metrics.LastStarted > 0 {
			metrics.TotalUptime += now - metrics.LastStarted
		}
		metrics.LastStarted = 0
	}

	return m.metrics.Resolve(newEndPoint, metrics)
}

func (m *ManagerMetrics) ResolveRequest(owner string, endPoint *mock_domain.EndPoint, response *mock_domain.Response, latency int64) *mock_domain.Metrics {
	if endPoint.Owner != owner {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, ok := m.metrics.Find(endPoint)
	if !ok {
		metrics = mock_domain.EmptyMetrics(endPoint)
	}

	key := fmt.Sprintf("%s-%d", response.Name, response.Timestamp)

	metrics.TotalRequests += 1
	metrics.CountResponses[key] += 1

	if metrics.MaxLatency < latency {
		metrics.MaxLatency = latency
	}

	if metrics.MinLatency > latency {
		metrics.MinLatency = latency
	}

	metrics.AvgLatency = metrics.AvgLatency + (float64(latency)-metrics.AvgLatency)/float64(metrics.TotalRequests)

	return m.metrics.Resolve(endPoint, metrics)
}

func (m *ManagerMetrics) Delete(owner string, endPoint *mock_domain.EndPoint) *mock_domain.Metrics {
	if endPoint.Owner != owner {
		return nil
	}

	metrics, ok := m.metrics.Find(endPoint)
	if !ok {
		return nil
	}

	return m.metrics.Delete(endPoint, metrics)
}
