package command

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	cmd_app "github.com/Rafael24595/go-api-core/src/commons/command/apps/app"
	cmd_log "github.com/Rafael24595/go-api-core/src/commons/command/apps/log"
	cmd_snapshot "github.com/Rafael24595/go-api-core/src/commons/command/apps/snapshot"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const Command apps.SnapshotFlag = "cmd"

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

var refRoot = apps.CommandReference{
	Flag:        Command,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_CMD_HELP),
}

var refApps = []apps.CommandApplication{
	cmd_app.App,
	cmd_log.App,
	cmd_snapshot.App,
}

func findApps() []apps.CommandReference {
	actions := make([]apps.CommandReference, len(refApps)+1)
	actions[0] = refRoot

	for i := 1; i < len(refApps)+1; i++ {
		actions[i] = refApps[i-1].CommandReference
	}

	return actions
}

func findApp(flag apps.SnapshotFlag) *apps.CommandApplication {
	for _, meta := range refApps {
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

	if position >= len(refApps) {
		position = InitialStep
	}

	coincidences, cursor, position := comp(*head, position, refApps)
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

func comp(head string, position int, actions []apps.CommandApplication) ([]apps.CommandApplication, *apps.CommandApplication, int) {
	var cursor *apps.CommandApplication

	cache := make(map[apps.SnapshotFlag]int)
	coincidences := make([]apps.CommandApplication, 0)

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
	if head == string(Command) || head == FLAG_CMD_HELP {
		return root(user, cmd)
	}

	action := findApp(apps.SnapshotFlag(head))
	if action == nil {
		return "", fmt.Errorf("unknown command %q", head)
	}

	return action.Exec(user, cmd)
}

func root(_ string, cmd *collection.Vector[string]) (string, error) {
	if cmd.Size() == 0 {
		return help(), nil
	}

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch *flag {
		case FLAG_CMD_HELP:
			return help(), nil
		default:
			return fmt.Sprintf("Unrecognized command flag: %s", *flag), nil
		}
	}

	return "", nil
}

func help() string {
	title := "Available cmd applications:\n"
	return apps.RunHelp(title, findApps())
}
