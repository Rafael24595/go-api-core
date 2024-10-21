package response

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type MemoryCommand struct {
	mu    sync.Mutex
	query IRepositoryQuery
	file  repository.IFileManager[domain.Response]
}

func NewMemoryCommand(query IRepositoryQuery) *MemoryCommand {
	return &MemoryCommand{
		query: query,
		file:  query.fileManager(),
	}
}

func (r *MemoryCommand) Insert(response domain.Response) *domain.Response {
	cursor, responses := r.query.insert(response)
	go r.write(responses)

	return &cursor
}

func (r *MemoryCommand) Delete(response domain.Response) *domain.Response {
	cursor, responses := r.query.delete(response)
	go r.write(responses)
	return &cursor
}

func (r *MemoryCommand) DeleteOptions(options repository.FilterOptions[domain.Response]) []string {
	ids, responses := r.query.deleteOptions(options, func(r domain.Response) string {
		return r.Id
	})
	go r.write(responses)
	return ids
}

func (r *MemoryCommand) write(responses []any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.file.Write(responses)
	if err != nil {
		println(err.Error())
	}
}