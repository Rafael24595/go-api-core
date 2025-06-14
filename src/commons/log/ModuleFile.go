package log

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Rafael24595/go-api-core/src/commons/routine"
	utils_commons "github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository/utils"
)

const MODULE_FILE = "FILE"
const LOG_PATH = "./.log"

type moduleFile struct {
	filePath  string
	session   string
	timestamp int64
	pool      *routine.StreamPool[[]Record]
	formatter Formatter
}

func newModuleFile(session string, timestamp int64) *moduleFile {
	pool := routine.SyncStreamPool[[]Record](100).
		EnableAutoDrain().
		Make()
	return &moduleFile{
		filePath:  LOG_PATH,
		session:   session,
		timestamp: timestamp,
		pool:      pool,
		formatter: Formatter{},
	}
}

func (m *moduleFile) Name() string {
	return MODULE_FILE
}

func (m *moduleFile) Vector(records []Record) []Record {
	result := m.pool.Submit(func(ctx context.Context) ([]Record, error) {
		return m.write(records)
	})

	if !result {
		fmt.Println("Cannot write file pool buffer is full")
	}

	return records
}

func (m *moduleFile) Record(record *Record, throwPanic bool) *Record {
	if throwPanic {
		message := m.formatter.Format(*record)
		panic(message)
	}
	return record
}

func (m *moduleFile) write(records []Record) ([]Record, error) {
	jsonData, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
	}

	name := fmt.Sprintf("log-%s-%s", m.session, utils_commons.FormatMillisecondsCompact(m.timestamp))
	path := fmt.Sprintf("%s/%s.json", m.filePath, name)
	err = utils.WriteFile(path, string(jsonData))
	if err != nil {
		fmt.Println(err.Error())
	}

	return records, nil
}
