package repository

import (
	"encoding/json"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

type ManagerJsonFile[T IStructure] struct {
	builder func() *T
	path string
}

func NewManagerJsonFile[T IStructure](builder func() *T, path string) *ManagerJsonFile[T] {
	return &ManagerJsonFile[T]{
		builder: builder,
		path: path,
	}
}

func (m *ManagerJsonFile[T]) Read() (map[string]T, error) {
	buffer, err := utils.ReadFile(m.path)
	if err != nil {
		return nil, err
	}

	if len(buffer) == 0 {
		return make(map[string]T), nil
	}

	var items []T
	err = json.Unmarshal(buffer, &items)
	if err != nil {
		return nil, err
	}

	return collection.Mapper(items, func(r T) string {
		return r.PersistenceId()
	}).Collect(), nil
}

func (m *ManagerJsonFile[T]) Write(items []any) error {
	jsonData, err := json.Marshal(items)
	if err != nil {
		return err
	}

	return utils.WriteFile(m.path, string(jsonData))
}