package request

import (
	"encoding/json"
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

type MemoryCommand struct {
	mu    sync.Mutex
	query IRepositoryQuery
	path  string
}

func NewMemoryCommand(query IRepositoryQuery) *MemoryCommand {
	return &MemoryCommand{
		query: query,
		path:  query.filePath(),
	}
}

func (r *MemoryCommand) Insert(request domain.Request) *domain.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, requests := r.query.insert(request)
	r.write(requests)
	return &cursor
}

func (r *MemoryCommand) Delete(request domain.Request) *domain.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, requests := r.query.delete(request)
	r.write(requests)
	return &cursor
}

func (r *MemoryCommand) DeleteOptions(options repository.FilterOptions[domain.Request]) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids, requests := r.query.deleteOptions(options, func(r domain.Request) string {
		return r.Id
	})

	r.write(requests)

	return ids
}

func (r *MemoryCommand) write(requests []any) error {
	jsonData, err := json.Marshal(requests)
	if err != nil {
		return err
	}

	return utils.WriteFile(r.path, string(jsonData))
}
