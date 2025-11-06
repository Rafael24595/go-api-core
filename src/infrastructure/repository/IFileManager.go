package repository

type IFileManager[T IStructure] interface {
	Read() (map[string]T, error)
	Write(items []T) error
	unmarshal(buffer []byte) (map[string]T, error)
	marshal(items []T) ([]byte, error)
}
