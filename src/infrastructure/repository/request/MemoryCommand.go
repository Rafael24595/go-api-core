package request

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type MemoryCommand struct {
	mu    sync.Mutex
	query IRepositoryQuery
	file  IFileManager
}

func NewMemoryCommand(query IRepositoryQuery) *MemoryCommand {
	return &MemoryCommand{
		query: query,
		file:  query.fileManager(),
	}
}

func (r *MemoryCommand) Insert(request domain.Request) *domain.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, requests := r.query.insert(request)
	r.file.Write(requests)
	return &cursor
}

func (r *MemoryCommand) Delete(request domain.Request) *domain.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, requests := r.query.delete(request)
	r.file.Write(requests)
	return &cursor
}

func (r *MemoryCommand) DeleteOptions(options repository.FilterOptions[domain.Request]) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids, requests := r.query.deleteOptions(options, func(r domain.Request) string {
		return r.Id
	})

	r.file.Write(requests)

	return ids
}