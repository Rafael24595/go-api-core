package response

import (
	"sync"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/csvt_translator"
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
		path:  DEFAULT_FILE_PATH,
	}
}

func (r *MemoryCommand) Insert(response domain.Response) *domain.Response {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, responses := r.query.insert(response)
	r.write(responses)
	return &cursor
}

func (r *MemoryCommand) Delete(response domain.Response) *domain.Response {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, responses := r.query.delete(response)
	r.write(responses)
	return &cursor
}

func (r *MemoryCommand) write(responses []any) error {
	csvt := csvt_translator.NewSerializer().
		Serialize(responses...)

	return utils.WriteFile(r.path, csvt)
}
