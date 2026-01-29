package mock

type RepositoryMetrics interface {
	Find(endPoint *EndPoint) (*Metrics, bool)
	Resolve(endPoint *EndPoint, metrics *Metrics) *Metrics
	Delete(endPoint *EndPoint, metrics *Metrics) *Metrics
}
