package cmd_app

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-collections/collection"
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

func exec(user string, cmd *collection.Vector[string]) (string, error) {
	if cmd.Size() == 0 {
		return help(), nil
	}

	messages := make([]string, 0)

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch *flag {
		case FLAG_HELP:
			return help(), nil
		case FLAG_VERSION:
			return version(), nil
		default:
			return fmt.Sprintf("Unrecognized command flag: %s", *flag), nil
		}
	}

	return strings.Join(messages, ", "), nil
}

func help() string {
	title := fmt.Sprintf("Available %s actions:\n", Command)
	return apps.RunHelp(title, refs)
}

func version() string {
	config := configuration.Instance()
	project := config.Project

	buffer := make([]string, 0)

	buffer = append(buffer, fmt.Sprintf("%s %s", project.Name, project.Version))

	return strings.Join(buffer, "\n")
}
