package apps

import (
	"fmt"
	"strings"
)

func RunHelp(title string, actions []CommandReference) string {
	result := make([]string, 0)
	result = append(result, title)
	for _, a := range actions {
		result = append(result, fmt.Sprintf(" %s: %s", a.Flag, a.Description))
		result = append(result, fmt.Sprintf("  Example: %s\n", a.Example))
	}
	return strings.Join(result, "\n")
}
