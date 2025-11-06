package repository

type IFileManager[T IStructure] interface {
	Read() (map[string]T, error)
	Write(items []T) error
}
