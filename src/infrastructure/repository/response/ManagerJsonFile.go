package response

import (
	"encoding/json"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

const (
	JSON_HISTORIC_FILE_PATH string = "./db/collection_response_historic.json"
	JSON_PERSISTED_FILE_PATH string = "./db/collection_response_persisted.json"
)

type ManagerJsonFile struct {
	path string
}

func NewManagerJsonFile(path string) *ManagerJsonFile {
	return &ManagerJsonFile{
		path: path,
	}
}

func (m *ManagerJsonFile) Read() (map[string]domain.Response, error) {
	buffer, err := utils.ReadFile(m.path)
	if err != nil {
		return nil, err
	}

	if len(buffer) == 0 {
		return make(map[string]domain.Response), nil
	}

	var requests []domain.Response
	err = json.Unmarshal(buffer, &requests)
	if err != nil {
		return nil, err
	}
	
	return collection.Mapper(requests, func(r domain.Response) string {
		return r.Id
	}).Collect(), nil
}

func (m *ManagerJsonFile) Write(responses []any) error {
	jsonData, err := json.Marshal(responses)
	if err != nil {
		return err
	}

	return utils.WriteFile(m.path, string(jsonData))
}