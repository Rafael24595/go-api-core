package command

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

type SnapshotFlag string

const (
	CMD      SnapshotFlag = "cmd"
	LOG      SnapshotFlag = "log"
	SNAPSHOT SnapshotFlag = "snpsh"
)

const (
	FLAG_CMD_HELP = "-h"
)

const InitialStep = -1

type CompleteHelp struct {
	Message     string
	Application string
	Position    int
	Lenght      int
}

func emptyCompleteHelp() *CompleteHelp {
	return &CompleteHelp{
		Message:     "",
		Application: "",
		Position:    InitialStep,
		Lenght:      0,
	}
}

type commandReference struct {
	Flag        SnapshotFlag
	Name        string
	Description string
	Example     string
}

type commandApplication struct {
	commandReference
	Exec func(user string, cmd *collection.Vector[string]) (string, error)
	Help func() string
}

var cmdRoot = commandReference{
	Flag:        CMD,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     `cmd -h`,
}

var cmdLog = commandApplication{
	commandReference: commandReference{
		Flag:        LOG,
		Name:        "Log",
		Description: "Manages log records",
		Example:     "log -h",
	},
	Exec: logg,
	Help: runLogHelp,
}

var cmdSnapshot = commandApplication{
	commandReference: commandReference{
		Flag:        SNAPSHOT,
		Name:        "Snapshot",
		Description: "Manages in-memory persistence snapshots",
		Example:     "snpsh -h",
	},
	Exec: snapshot,
	Help: runSnapshotHelp,
}

var cmdActions = []commandApplication{
	cmdLog,
	cmdSnapshot,
}

func findActions() []commandReference {
	actions := make([]commandReference, len(cmdActions)+1)
	actions[0] = cmdRoot

	for i := 1; i < len(cmdActions); i++ {
		actions[i] = cmdActions[i].commandReference
	}

	return actions
}

func findActionMetadata(flag SnapshotFlag) *commandApplication {
	for _, meta := range cmdActions {
		if meta.Flag == flag {
			return &meta
		}
	}
	return nil
}

func Comp(user, command string, position int) (*CompleteHelp, error) {
	raw, err := utils.SplitCommand(command)
	if err != nil {
		return nil, err
	}

	cmd := collection.VectorFromList(raw)

	head, ok := cmd.Shift()
	if !ok {
		return &CompleteHelp{
			Message:     "",
			Application: command,
			Position:    InitialStep,
			Lenght:      0,
		}, nil
	}

	if position >= len(cmdActions) {
		position = InitialStep
	}

	coincidences, cursor, position := comp(*head, position, cmdActions)
	if cursor == nil {
		return emptyCompleteHelp(), nil
	}

	message := ""
	if len(coincidences) > 1 {
		buffer := make([]string, len(coincidences))
		for i := range coincidences {
			buffer[i] = string(coincidences[i].Flag)
		}
		message = strings.Join(buffer, " ")
	} else if len(coincidences) == 1 && string(cursor.Flag) == *head {
		message = cursor.Help()
	}

	return &CompleteHelp{
		Message:     message,
		Application: string(cursor.Flag),
		Position:    position,
		Lenght:      len(coincidences),
	}, nil
}

func comp(head string, position int, actions []commandApplication) ([]commandApplication, *commandApplication, int) {
	var cursor *commandApplication

	cache := make(map[SnapshotFlag]int)
	coincidences := make([]commandApplication, 0)

	for i, v := range actions {
		if !strings.HasPrefix(string(v.Flag), head) {
			continue
		}

		cache[v.Flag] = i
		coincidences = append(coincidences, v)

		if cursor == nil && i > position {
			cursor = &v
		}
	}

	if cursor == nil {
		if len(coincidences) > 0 {
			return coincidences, &coincidences[0], cache[coincidences[0].Flag]
		}
		return coincidences, nil, InitialStep
	}

	return coincidences, cursor, cache[cursor.Flag]
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

	return exec(user, *head, cmd)
}

func exec(user, head string, cmd *collection.Vector[string]) (string, error) {
	if head == string(CMD) || head == FLAG_CMD_HELP {
		return root(user, cmd)
	}

	action := findActionMetadata(SnapshotFlag(head))
	if action == nil {
		return "", fmt.Errorf("unknown command %q", head)
	}

	return action.Exec(user, cmd)
}

func root(_ string, cmd *collection.Vector[string]) (string, error) {
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
	return runHelp(title, findActions())
}

func runHelp(title string, actions []commandReference) string {
	result := make([]string, 0)
	result = append(result, title)
	for _, a := range actions {
		result = append(result, fmt.Sprintf(" %s: %s", a.Flag, a.Description))
		result = append(result, fmt.Sprintf("  Example: %s\n", a.Example))
	}
	return strings.Join(result, "\n")
}
