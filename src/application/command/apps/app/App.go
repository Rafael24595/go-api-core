package cmd_app

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/application/command/apps"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
)

const Command apps.SnapshotFlag = "app"

const (
	FLAG_HELP    = "-h"
	FLAG_VERSION = "-v"
)

var App = apps.CommandApplication{
	CommandReference: apps.CommandReference{
		Flag:        Command,
		Name:        "go-api",
		Description: "Retrieves the project metadata",
		Example:     refHelp.Example,
	},
	Exec: exec,
	Help: help,
}

var refs = []apps.CommandReference{
	refHelp,
	refVersion,
}

var refHelp = apps.CommandReference{
	Flag:        FLAG_HELP,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_HELP),
}

var refVersion = apps.CommandReference{
	Flag:        FLAG_VERSION,
	Name:        "Version",
	Description: "Show the project version.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_VERSION),
}

func exec(request *apps.CmdExecRequest) *apps.CmdExecResult {
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
		case FLAG_VERSION:
			return version()
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return apps.EmptyResult()
}

func help() *apps.CmdExecResult {
	title := fmt.Sprintf("Available %s actions:\n", Command)
	return apps.RunHelp(title, refs)
}

func version() *apps.CmdExecResult {
	config := configuration.Instance()
	project := config.Project

	buffer := make([]string, 0)

	buffer = append(buffer, fmt.Sprintf("%s %s", project.Name, project.Version))

	return apps.NewResult(strings.Join(buffer, "\n"))
}
