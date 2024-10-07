package request

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/csvt_translator"
)

type MemoryCommand struct {
	mu    sync.Mutex
	query RepositoryQuery
	path  string
}

func NewMemoryCommand(query RepositoryQuery) *MemoryCommand {
	return &MemoryCommand{
		query: query,
		path:  DEFAULT_FILE_PATH,
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

func (r *MemoryCommand) write(requests []any) error {
	csvt := csvt_translator.NewSerializer().
		Serialize(requests...)

	return writeFile(r.path, csvt)
}
