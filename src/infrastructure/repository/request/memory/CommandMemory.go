package memory

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/csvt_translator"
)

type CommandMemory struct {
	mu    sync.Mutex
	query *QueryMemory
	path  string
}

func NewCommandMemory(query *QueryMemory) *CommandMemory {
	return &CommandMemory{
		query: query,
		path:  DEFAULT_FILE_PATH,
	}
}

func (r *CommandMemory) Insert(request domain.Request) *domain.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, requests := r.query.insert(request)
	r.write(requests)
	return &cursor
}

func (r *CommandMemory) write(requests []any) error {
	csvt := csvt_translator.NewSerializer().
		Serialize(requests...)

	return writeFile(r.path, csvt)
}
