package repository

import (
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

	return UnmarshalCsvt[T](buffer)
}

func (m *ManagerCsvtFile[T]) Write(items []T) error {
	result, err := MarshalCsvt(items)
	if err != nil {
		return err
	}

	return utils.WriteFileSafe(m.path, string(result))
}
