package request

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/csvt_translator"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

const (
	CSVT_HISTORIC_FILE_PATH string = "./db/collection_request_historic.csvt"
	CSVT_PERSISTED_FILE_PATH string = "./db/collection_request_persisted.csvt"
)

type ManagerCsvtFile struct {
	path string
}

func NewManagerCsvtFile(path string) *ManagerCsvtFile {
	return &ManagerCsvtFile{
		path: path,
	}
}

func (m *ManagerCsvtFile) Read() (map[string]domain.Request, error) {
	buffer, err := utils.ReadFile(m.path)
	if err != nil {
		return nil, err
	}

	if len(buffer) == 0 {
		return make(map[string]domain.Request), nil
	}

	deserializer, err := csvt_translator.NewDeserialzer(string(buffer))

	requests := map[string]domain.Request{}

	iterator := deserializer.Iterate()
	for iterator.Next() {
		request := &domain.Request{}
		_ , err := iterator.Deserialize(request)
		if err != nil {
			return nil, err
		}
		requests[request.Id] = *request
	}

	return requests, nil
}

func (m *ManagerCsvtFile) Write(requests []any) error {
	csvt := csvt_translator.NewSerializer().
		Serialize(requests...)

	return utils.WriteFile(m.path, csvt)
}