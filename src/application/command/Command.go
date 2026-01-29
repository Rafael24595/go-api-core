package command

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Rafael24595/go-api-core/src/application/command/apps"
	cmd_app "github.com/Rafael24595/go-api-core/src/application/command/apps/app"
	cmd_log "github.com/Rafael24595/go-api-core/src/application/command/apps/log"
	cmd_repo "github.com/Rafael24595/go-api-core/src/application/command/apps/repo"
	cmd_snapshot "github.com/Rafael24595/go-api-core/src/application/command/apps/snapshot"
	cmd_user "github.com/Rafael24595/go-api-core/src/application/command/apps/user"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const Command apps.SnapshotFlag = "cmd"

const (
	FLAG_HELP = "-h"
)

const InitialStep = -1

type CmdAuxApp struct {
	Order int
	Flag  string
	Help  string
}

func auxAppToApp(aux ...CmdAuxApp) []apps.CommandApplication {
	app := make([]apps.CommandApplication, len(aux))

	sort.Slice(aux, func(i, j int) bool {
		return aux[i].Order < aux[j].Order
	})

	for i, v := range aux {
		app[i] = apps.CommandApplication{
			CommandReference: apps.CommandReference{
				Flag: apps.SnapshotFlag(v.Flag),
			},
			Exec: func(request *apps.CmdExecRequest) *apps.CmdExecResult {
				return apps.EmptyResult()
			},
			Help: func() *apps.CmdExecResult {
				return apps.NewResult(v.Help)
			},
		}
	}
	return app
}

type CmdCompResult struct {
	Message     string
	Application string
	Position    int
	Lenght      int
}

func emptyCompleteHelp() *CmdCompResult {
	return &CmdCompResult{
		Message:     "",
		Application: "",
		Position:    InitialStep,
		Lenght:      0,
	}
}

var App = apps.CommandApplication{
	CommandReference: apps.CommandReference{
		Flag:        Command,
		Name:        "go-api",
		Description: "Retrieves the project metadata",
		Example:     refHelp.Example,
	},
	Exec: root,
	Help: help,
}

var refHelp = apps.CommandReference{
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
	cmd_repo.App,
}

func findApps() []apps.CommandReference {
	actions := make([]apps.CommandReference, len(refApps)+1)
	actions[0] = refHelp

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

func Comp(user, command string, position int, auxApps ...CmdAuxApp) (*CmdCompResult, error) {
	raw, err := utils.SplitCommand(command)
	if err != nil {
		return nil, err
	}

	cmd := collection.VectorFromList(raw)

	head, ok := cmd.Shift()
	if !ok {
		return &CmdCompResult{
			Message:     "",
			Application: command,
			Position:    InitialStep,
			Lenght:      0,
		}, nil
	}

	apps := []apps.CommandApplication{App}

	apps = append(apps, refApps...)
	apps = append(apps, auxAppToApp(auxApps...)...)

	if position >= len(apps) {
		position = InitialStep
	}

	coincidences, cursor, position := comp(head, position, apps)
	if cursor == nil {
		return emptyCompleteHelp(), nil
	}

	message := ""
	if len(coincidences) > 1 {
		message = strings.Join(coincidences, " ")
	} else if len(coincidences) == 1 && string(cursor.Flag) == head {
		message = cursor.Help().Output
	}

	return &CmdCompResult{
		Message:     message,
		Application: string(cursor.Flag),
		Position:    position,
		Lenght:      len(coincidences),
	}, nil
}

func comp(head string, position int, actions []apps.CommandApplication) ([]string, *apps.CommandApplication, int) {
	var cursor *apps.CommandApplication

	cache := make(map[apps.SnapshotFlag]int)
	apps := make([]apps.CommandApplication, 0)

	for i, v := range actions {
		if !strings.HasPrefix(string(v.Flag), head) {
			continue
		}

		cache[v.Flag] = i
		apps = append(apps, v)

		if cursor == nil && i > position {
			cursor = &v
		}
	}

	flags := make([]string, len(apps))
	for i, v := range apps {
		flags[i] = string(v.Flag)
	}

	if cursor != nil {
		return flags, cursor, cache[cursor.Flag]
	}

	if len(apps) > 0 {
		return flags, &apps[0], cache[apps[0].Flag]
	}

	return flags, nil, InitialStep
}

func Exec(user, input string) *apps.CmdExecResult {
	raw, err := utils.SplitCommand(input)
	if err != nil {
		return apps.ErrorResult(err)
	}

	cmd := collection.VectorFromList(raw)

	app, ok := cmd.Shift()
	if !ok {
		return apps.EmptyResult()
	}

	request := &apps.CmdExecRequest{
		User:    user,
		Input:   input,
		Command: cmd,
	}

	return exec(app, request)
}

func exec(app string, request *apps.CmdExecRequest) *apps.CmdExecResult {
	if app == string(Command) || app == FLAG_HELP {
		return root(request)
	}

	action := findApp(apps.SnapshotFlag(app))
	if action == nil {
		return apps.NewResultf("unknown command %q", app)
	}

	return action.Exec(request)
}

func root(request *apps.CmdExecRequest) *apps.CmdExecResult {
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

func help() *apps.CmdExecResult {
	title := fmt.Sprintf("Available %s actions:\n", Command)
	return apps.RunHelp(title, findApps())
}
