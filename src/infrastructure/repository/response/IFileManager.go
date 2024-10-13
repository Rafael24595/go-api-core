package response

import "github.com/Rafael24595/go-api-core/src/domain"

type IFileManager interface {
	Read() (map[string]domain.Response, error)
	Write(responses []any) error
}