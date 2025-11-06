package repository

import (
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
	"github.com/Rafael24595/go-csvt/csvt"
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

	return m.unmarshal(buffer)
}

func (m *ManagerCsvtFile[T]) Write(items []T) error {
	result, err := m.marshal(items)
	if err != nil {
		return err
	}

	return utils.WriteFile(m.path, string(result))
}

func (m *ManagerCsvtFile[T]) unmarshal(buffer []byte) (map[string]T, error) {
	if len(buffer) == 0 {
		return make(map[string]T), nil
	}

	var vector []T
	err := csvt.Unmarshal(buffer, &vector)
	if err != nil {
		return nil, err
	}

	items := map[string]T{}
	for _, v := range vector {
		items[v.PersistenceId()] = v
	}

	return items, nil
}

func (m *ManagerCsvtFile[T]) marshal(snapshot []T) ([]byte, error) {
	items := make([]any, 0)
	for v := range snapshot {
		items = append(items, v)
	}

	return csvt.Marshal(items...)
}
