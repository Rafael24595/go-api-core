package apps

import "github.com/Rafael24595/go-collections/collection"

type SnapshotFlag string

type CommandReference struct {
	Flag        SnapshotFlag
	Name        string
	Description string
	Example     string
}

type CommandApplication struct {
	CommandReference
	Exec func(user string, cmd *collection.Vector[string]) (string, error)
	Help func() string
}
