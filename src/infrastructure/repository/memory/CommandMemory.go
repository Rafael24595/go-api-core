package memory

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/memory/parser"
)

type CommandMemory struct {
	mu    sync.Mutex
	query *QueryMemory
}

func NewCommandMemory(query *QueryMemory) *CommandMemory {
	return &CommandMemory{
		query: query,
	}
}

func (r *CommandMemory) Insert(request domain.Request) *domain.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, requests := r.query.insert(request)
	r.write(requests)
	return &cursor;
}

func (r *CommandMemory) write(requests []any) error {
	csvt := parser.NewSerializer().
		Serialize(requests...)

	return writeFile(DEFAULT_FILE_PATH, csvt)
}