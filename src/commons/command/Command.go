package command

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const (
	CMD      = "cmd"
	LOG      = "log"
	SNAPSHOT = "snpsh"
)

const (
	FLAG_CMD_HELP = "-h"
)

type commandAction struct {
	Flag        string
	Name        string
	Description string
	Example     string
}

var cmdHelp = commandAction{
	Flag:        FLAG_CMD_HELP,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     `cmd -h`,
}

var cmdLog = commandAction{
	Flag:        LOG,
	Name:        "Log",
	Description: "Manages log records",
	Example:     "log -h",
}

var cmdSnapshot = commandAction{
	Flag:        SNAPSHOT,
	Name:        "Snapshot",
	Description: "Manages in-memory persistence snapshots",
	Example:     "snpsh -h",
}

var cmdActions = []commandAction{
	cmdHelp,
	cmdLog,
	cmdSnapshot,
}

func Exec(user, command string) (string, error) {
	raw, err := utils.SplitCommand(command)
	if err != nil {
		return "", err
	}

	cmd := collection.VectorFromList(raw)

	head, ok := cmd.Shift()
	if !ok {
		return "", nil
	}

	return run(user, *head, cmd)
}

func run(user, head string, cmd *collection.Vector[string]) (string, error) {
	switch head {
	case CMD:
		return root(cmd)
	case LOG:
		return logg(user, cmd)
	case SNAPSHOT:
		return snapshot(cmd)
	}
	return "", fmt.Errorf("unknown command %q", head)
}

func root(cmd *collection.Vector[string]) (string, error) {
	if cmd.Size() == 0 {
		return runCmdHelp(), nil
	}

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch *flag {
		case FLAG_CMD_HELP:
			return runCmdHelp(), nil
		default:
			return fmt.Sprintf("Unrecognized command flag: %s", *flag), nil
		}
	}

	return "", nil
}

func runCmdHelp() string {
	title := "Available cmd applications:\n"
	return runHelp(title, cmdActions)
}

func runHelp(title string, actions []commandAction) string {
	result := make([]string, 0)
	result = append(result, title)
	for _, a := range actions {
		result = append(result, fmt.Sprintf(" %s: %s", a.Flag, a.Description))
		result = append(result, fmt.Sprintf("  Example: %s\n", a.Example))
	}
	return strings.Join(result, "\n")
}
