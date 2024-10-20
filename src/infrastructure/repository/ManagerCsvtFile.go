package repository

import (
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/csvt_translator"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

type ManagerCsvtFile[T IStructure] struct {
	builder func() *T
	path string
}

func NewManagerCsvtFile[T IStructure](builder func() *T, path string) *ManagerCsvtFile[T] {
	return &ManagerCsvtFile[T]{
		builder: builder,
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

	deserializer, err := csvt_translator.NewDeserialzer(string(buffer))

	items := map[string]T{}

	iterator := deserializer.Iterate()
	for iterator.Next() {
		item := m.builder()
		_ , err := iterator.Deserialize(item)
		if err != nil {
			return nil, err
		}
		items[(*item).PersistenceId()] = *item
	}

	return items, nil
}

func (m *ManagerCsvtFile[T]) Write(items []any) error {
	csvt := csvt_translator.NewSerializer().
		Serialize(items...)

	return utils.WriteFile(m.path, csvt)
}