package request

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type MemoryCommand struct {
	mu    sync.Mutex
	query IRepositoryQuery
	file  repository.IFileManager[domain.Request]
}

func NewMemoryCommand(query IRepositoryQuery) *MemoryCommand {
	return &MemoryCommand{
		query: query,
		file:  query.fileManager(),
	}
}

func (r *MemoryCommand) Insert(request domain.Request) *domain.Request {
	cursor, requests := r.query.insert(request)
	go r.write(requests)
	return &cursor
}

func (r *MemoryCommand) Delete(request domain.Request) *domain.Request {
	cursor, requests := r.query.delete(request)
	go r.write(requests)
	return &cursor
}

func (r *MemoryCommand) DeleteOptions(options repository.FilterOptions[domain.Request]) []string {
	ids, requests := r.query.deleteOptions(options, func(r domain.Request) string {
		return r.Id
	})
	go r.write(requests)
	return ids
}

func (r *MemoryCommand) write(requests []any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.file.Write(requests)
	if err != nil {
		println(err.Error())
	}
}