package mock

import (
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/mock/swr"
	"github.com/Rafael24595/go-collections/collection"
)

type EndPoint struct {
	Id        string            `json:"_id"`
	Status    bool              `json:"status"`
	Order     int               `json:"order"`
	Timestamp int64             `json:"timestamp"`
	Modified  int64             `json:"modified"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Path      string            `json:"path"`
	Responses []Response        `json:"responses"`
	Safe      bool              `json:"safe"`
	Owner     string            `json:"owner"`
}

func (r EndPoint) DefaultResponse() Response {
	return *defaultResponse()
}

func (r EndPoint) PersistenceId() string {
	return r.Id
}

type EndPointLite struct {
	Id        string            `json:"_id"`
	Status    bool              `json:"status"`
	Order     int               `json:"order"`
	Timestamp int64             `json:"timestamp"`
	Modified  int64             `json:"modified"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Path      string            `json:"path"`
	Responses []string          `json:"responses"`
	Safe      bool              `json:"safe"`
	Owner     string            `json:"owner"`
}

func LiteFromEndPoint(endPoint *EndPoint) *EndPointLite {
	keys := collection.MapToVector(endPoint.Responses, func(r Response) string {
		return r.Name
	})

	return &EndPointLite{
		Id:        endPoint.Id,
		Status:    endPoint.Status,
		Order:     endPoint.Order,
		Timestamp: endPoint.Timestamp,
		Modified:  endPoint.Modified,
		Name:      endPoint.Name,
		Method:    endPoint.Method,
		Path:      endPoint.Path,
		Responses: keys.Collect(),
		Safe:      endPoint.Safe,
		Owner:     endPoint.Owner,
	}
}

type EndPointFull struct {
	Id        string            `json:"_id"`
	Status    bool              `json:"status"`
	Order     int               `json:"order"`
	Timestamp int64             `json:"timestamp"`
	Modified  int64             `json:"modified"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Path      string            `json:"path"`
	Responses []ResponseFull    `json:"responses"`
	Safe      bool              `json:"safe"`
	Owner     string            `json:"owner"`
}

func FullFromEndPoint(endPoint *EndPoint) (*EndPointFull, []error) {
	responses := make([]ResponseFull, len(endPoint.Responses))
	errs := make([]error, 0)

	opts := swr.UnmarshalOpts{Evalue: true}
	for i, v := range endPoint.Responses {
		result, resErrs := FromResponseWithOptions(v, opts)
		if len(errs) > 0 {
			errs = append(errs, resErrs...)
		}

		responses[i] = *result
	}

	return &EndPointFull{
		Id:        endPoint.Id,
		Status:    endPoint.Status,
		Order:     endPoint.Order,
		Timestamp: endPoint.Timestamp,
		Modified:  endPoint.Modified,
		Name:      endPoint.Name,
		Method:    endPoint.Method,
		Path:      endPoint.Path,
		Responses: responses,
		Safe:      endPoint.Safe,
		Owner:     endPoint.Owner,
	}, errs
}

func ToEndPointFromFull(endPoint *EndPointFull) (*EndPoint, []error) {
	responses := make([]Response, len(endPoint.Responses))

	opts := swr.MarshalOpts{Evalue: true}
	for i, v := range endPoint.Responses {
		result, errs := ToResponseWithOptions(v, opts)
		if len(errs) > 0 {
			return nil, errs
		}

		responses[i] = *result
	}

	return &EndPoint{
		Id:        endPoint.Id,
		Status:    endPoint.Status,
		Order:     endPoint.Order,
		Timestamp: endPoint.Timestamp,
		Modified:  endPoint.Modified,
		Name:      endPoint.Name,
		Method:    endPoint.Method,
		Path:      endPoint.Path,
		Responses: responses,
		Safe:      endPoint.Safe,
		Owner:     endPoint.Owner,
	}, make([]error, 0)
}

func FixEndPoints(owner string, endPoints []EndPoint) []EndPoint {
	coll := collection.VectorFromList(endPoints)

	coll.Sort(func(i, j EndPoint) bool {
		return i.Order < j.Order
	})

	coll.Map(func(i int, r EndPoint) EndPoint {
		r.Order = i
		return *FixEndPoint(owner, &r)
	})

	return coll.Collect()
}

func FixEndPoint(owner string, endPoint *EndPoint) *EndPoint {
	endPoint.Owner = owner

	if endPoint.Timestamp == 0 {
		endPoint.Timestamp = time.Now().UnixMilli()
	}

	endPoint.Modified = time.Now().UnixMilli()

	if endPoint.Name == "" {
		endPoint.Name = endPoint.Path
	}

	return endPoint
}
