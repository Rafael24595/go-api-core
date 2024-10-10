package request

import (
	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
)

type MemoryCommandManager struct {
	qHistoric IRepositoryQuery
	cHistoric IRepositoryCommand
	persisted IRepositoryCommand
}

func NewMemoryCommandManager(qHistoric IRepositoryQuery, cHistoric IRepositoryCommand, persisted IRepositoryCommand) *MemoryCommandManager {
	return &MemoryCommandManager{
		qHistoric: qHistoric,
		cHistoric: cHistoric,
		persisted: persisted,
	}
}

func (r *MemoryCommandManager) Insert(request domain.Request) *domain.Request {
	if request.Status == domain.Historic {
		result := /*go*/ r.insertHistoric(request)
		return result
	}
	return r.persisted.Insert(request)
}

func (r *MemoryCommandManager) insertHistoric(request domain.Request) *domain.Request {
	result := r.cHistoric.Insert(request)
	requests := r.qHistoric.FindAll()
	collection.FromList(requests).
		Sort(func(i, j domain.Request) bool {
			return i.Timestamp > j.Timestamp
		}).
		ForEach(func(i int, v domain.Request) {
			if i > 10 {
				r.cHistoric.Delete(v)
			}
		})
	return result
}