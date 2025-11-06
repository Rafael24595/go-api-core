package repository

import (
	"fmt"
	
	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-csvt/csvt"
)

func TryUnmarshal[T IStructure](frmt format.DataFormat, buffer []byte) (map[string]T, error) {
	switch frmt {
	case format.CSVT:
		return UnmarshalCsvt[T](buffer)
	default:
		return make(map[string]T), fmt.Errorf("unknown format type: %s", frmt)
	}
}

func TryMarshal[T IStructure](frmt format.DataFormat, snapshot []T) ([]byte, error) {
	switch frmt {
	case format.CSVT:
		return MarshalCsvt(snapshot)
	default:
		return make([]byte, 0), fmt.Errorf("unknown format type: %s", frmt)
	}
}

func UnmarshalCsvt[T IStructure](buffer []byte) (map[string]T, error) {
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

func MarshalCsvt[T IStructure](snapshot []T) ([]byte, error) {
	items := make([]any, 0)
	for _, v := range snapshot {
		items = append(items, v)
	}

	return csvt.Marshal(items...)
}
