package mock

type Metrics struct {
	EndPoint       string            `json:"end_point"`
	Timestamp      int64             `json:"timestamp"`
	Modified       int64             `json:"modified"`
	TotalUptime    int64             `json:"total_uptime"`
	LastStarted    int64             `json:"last_started"`
	TotalRequests  int64             `json:"total_requests"`
	CountResponses map[string]uint64 `json:"count_responses"`
	MinLatency     int64             `json:"min_latency"`
	MaxLatency     int64             `json:"max_latency"`
	AvgLatency     float64           `json:"avg_latency"`
}

func EmptyMetrics(endPoint *EndPoint) *Metrics {
	return &Metrics{
		EndPoint:       endPoint.Id,
		Timestamp:      0,
		Modified:       0,
		TotalUptime:    0,
		LastStarted:    0,
		TotalRequests:  0,
		CountResponses: make(map[string]uint64),
		MinLatency:     0,
		MaxLatency:     0,
		AvgLatency:     0,
	}
}

func (r Metrics) PersistenceId() string {
	return r.EndPoint
}
