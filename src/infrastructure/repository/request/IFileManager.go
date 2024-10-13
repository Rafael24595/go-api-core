package request

import "github.com/Rafael24595/go-api-core/src/domain"

type IFileManager interface {
	Read() (map[string]domain.Request, error)
	Write(requests []any) error
}