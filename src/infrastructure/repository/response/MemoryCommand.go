package response

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

func (r *MemoryCommand) Insert(response domain.Response) *domain.Response {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, responses := r.query.insert(response)
	r.file.Write(responses)
	return &cursor
}

func (r *MemoryCommand) Delete(response domain.Response) *domain.Response {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, responses := r.query.delete(response)
	r.file.Write(responses)
	return &cursor
}

func (r *MemoryCommand) DeleteOptions(options repository.FilterOptions[domain.Response]) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids, requests := r.query.deleteOptions(options, func(r domain.Response) string {
		return r.Id
	})

	r.file.Write(requests)

	return ids
}