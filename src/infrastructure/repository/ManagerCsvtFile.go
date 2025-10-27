package repository

import (
	"github.com/Rafael24595/go-csvt/csvt"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

type ManagerCsvtFile[T IStructure] struct {
	path string
}

func NewManagerCsvtFile[T IStructure](path string) *ManagerCsvtFile[T] {
	return &ManagerCsvtFile[T]{
		path: path,
	}
}

func (m *ManagerCsvtFile[T]) Read() (map[string]T, error) {
	buffer, err := utils.ReadFile(m.path)
	if err != nil {
		return nil, err
	}

	if len(buffer) == 0 {
		return make(map[string]T), nil
	}

	var vector []T
	err = csvt.Unmarshal(buffer, &vector)
	if err != nil {
		return nil, err
	}

	items := map[string]T{}
	for _, v := range vector {
		items[v.PersistenceId()] = v
	}

	return items, nil
}

func (m *ManagerCsvtFile[T]) Write(items []any) error {
	result, err := csvt.Marshal(items...)
	if err != nil {
		return err
	}

	return utils.WriteFile(m.path, string(result))
}