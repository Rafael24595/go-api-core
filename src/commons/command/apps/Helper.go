package apps

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

func ResolveValueCursor(cmd *collection.Vector[string]) (string, error) {
	value, cmd, err := TakeCursorValue(cmd, true)
	if err != nil {
		return "", err
	}

	if value == "" {
		return "", errors.New("value expected, but nothing found")
	}

	return value, nil
}

func ResolveKeyValueCursor(cmd *collection.Vector[string], sep string, required bool) (*utils.CmdTuple, error) {
	value, cmd, err := TakeCursorValue(cmd, required)
	if err != nil {
		return nil, err
	}

	if value == "" && !required {
		return nil, nil
	}

	cat, mes, ok := strings.Cut(value, sep)
	if !ok {
		return nil, fmt.Errorf("key value tuple expected, but '%s' found", value)
	}

	return utils.NewCmdTuple(cat, mes), nil
}

func ResolveChainCursor(cmd *collection.Vector[string], sep string) ([]string, error) {
	value, cmd, err := TakeCursorValue(cmd, true)
	if err != nil {
		return nil, err
	}
	return strings.Split(value, sep), nil
}

func TakeCursorValue(cmd *collection.Vector[string], required bool) (string, *collection.Vector[string], error) {
	next, _ := cmd.First()
	if strings.HasPrefix(next, "-") {
		if required {
			return "", cmd, fmt.Errorf("value expected, but flag '%s' found", next)
		}
		return "", cmd, nil
	}

	value, ok := cmd.Shift()
	if !ok {
		return "", cmd, nil
	}

	return value, cmd, nil
}

func RunHelp(title string, actions []CommandReference) *CmdResult {
	result := make([]string, 0)
	result = append(result, title)
	for _, a := range actions {
		result = append(result, fmt.Sprintf(" %s: %s", a.Flag, a.Description))
		result = append(result, fmt.Sprintf("  Example: %s\n", a.Example))
	}
	return NewResult(strings.Join(result, "\n"))
}
