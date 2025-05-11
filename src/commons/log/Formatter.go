package log

import (
	"fmt"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
)

type Formatter struct {
	//
}

func (f Formatter) Format(record Record) string {
	return fmt.Sprintf("%s - [%s]: %s", utils.FormatMilliseconds(record.Timestamp), record.Category, record.Message)
}