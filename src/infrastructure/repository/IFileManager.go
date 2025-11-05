package repository

type IFileManager[T IStructure] interface {
	Read() (map[string]T, error)
	Write(items []any) error
	marshal(items []any) ([]byte, error)
}