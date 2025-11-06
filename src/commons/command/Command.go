package command

import (
	"fmt"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const (
	SNAPSHOT = "snpsh"
)

func Exec(command string) (string, error) {
	raw, err := utils.SplitCommand(command)
	if err != nil {
		return "", err
	}

	cmd := collection.VectorFromList(raw)

	head, ok := cmd.Shift()
	if !ok {
		return "", nil
	}

	return run(*head, cmd)
}

func run(head string, cmd *collection.Vector[string]) (string, error) {
	switch head {
	case SNAPSHOT:
		return snapshot(cmd)
	}
	return "", fmt.Errorf("unknown command %q", head)
}
