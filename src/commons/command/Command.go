package command

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	cmd_app "github.com/Rafael24595/go-api-core/src/commons/command/apps/app"
	cmd_log "github.com/Rafael24595/go-api-core/src/commons/command/apps/log"
	cmd_snapshot "github.com/Rafael24595/go-api-core/src/commons/command/apps/snapshot"
	cmd_user "github.com/Rafael24595/go-api-core/src/commons/command/apps/user"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const Command apps.SnapshotFlag = "cmd"

const (
	FLAG_HELP = "-h"
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
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_HELP),
}

var refApps = []apps.CommandApplication{
	cmd_app.App,
	cmd_log.App,
	cmd_snapshot.App,
	cmd_user.App,
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

	coincidences, cursor, position := comp(head, position, refApps)
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
	} else if len(coincidences) == 1 && string(cursor.Flag) == head {
		message = cursor.Help().Output
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

func Exec(user, input string) *apps.CmdResult {
	raw, err := utils.SplitCommand(input)
	if err != nil {
		return apps.ErrorResult(err)
	}

	cmd := collection.VectorFromList(raw)

	app, ok := cmd.Shift()
	if !ok {
		return apps.EmptyResult()
	}

	request := &apps.CmdRequest{
		User:    user,
		Input:   input,
		Command: cmd,
	}

	return exec(app, request)
}

func exec(app string, request *apps.CmdRequest) *apps.CmdResult {
	if app == string(Command) || app == FLAG_HELP {
		return root(request)
	}

	action := findApp(apps.SnapshotFlag(app))
	if action == nil {
		return apps.NewResultf("unknown command %q", app)
	}

	return action.Exec(request)
}

func root(request *apps.CmdRequest) *apps.CmdResult {
	cmd := request.Command

	if cmd.Size() == 0 {
		return help()
	}

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_HELP:
			return help()
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return apps.EmptyResult()
}

func help() *apps.CmdResult {
	title := fmt.Sprintf("Available %s actions:\n", Command)
	return apps.RunHelp(title, findApps())
}
