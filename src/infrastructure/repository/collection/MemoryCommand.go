package collection

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
)

type MemoryCommand struct {
	mu    sync.Mutex
	query IRepositoryQuery
	file  repository.IFileManager[domain.Collection]
}

func NewMemoryCommand(query IRepositoryQuery) *MemoryCommand {
	return &MemoryCommand{
		query: query,
		file:  query.fileManager(),
	}
}

func (r *MemoryCommand) Insert(collection domain.Collection) *domain.Collection {
	cursor, collections := r.query.insert(collection)
	go r.write(collections)

	return &cursor
}

func (r *MemoryCommand) Delete(collection domain.Collection) *domain.Collection {
	cursor, collections := r.query.delete(collection)
	go r.write(collections)
	return &cursor
}

func (r *MemoryCommand) DeleteOptions(options repository.FilterOptions[domain.Collection]) []string {
	ids, collections := r.query.deleteOptions(options, func(r domain.Collection) string {
		return r.Id
	})
	go r.write(collections)
	return ids
}

func (r *MemoryCommand) write(collections []any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.file.Write(collections)
	if err != nil {
		println(err.Error())
	}
}