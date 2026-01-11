package cmd_log

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const Command apps.SnapshotFlag = "log"

const (
	FLAG_HELP = "-h"
	FLAG_LIST = "-l"
	FLAG_PUSH = "-p"
)

var App = apps.CommandApplication{
	CommandReference: apps.CommandReference{
		Flag:        Command,
		Name:        "Log",
		Description: "Manages log records",
		Example:     refHelp.Example,
	},
	Exec: exec,
	Help: help,
}

var refs = []apps.CommandReference{
	refHelp,
	refList,
	refPush,
}

var refHelp = apps.CommandReference{
	Flag:        FLAG_HELP,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_HELP),
}

var refList = apps.CommandReference{
	Flag:        FLAG_LIST,
	Name:        "List",
	Description: "Displays the list of log records.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_LIST),
}

var refPush = apps.CommandReference{
	Flag:        FLAG_PUSH,
	Name:        "Push",
	Description: "Insert a new log record.",
	Example:     fmt.Sprintf(`%s %s ${category}=${message}`, Command, FLAG_PUSH),
}

func exec(user string, cmd *collection.Vector[string]) (string, error) {
	if cmd.Size() == 0 {
		return help(), nil
	}

	pushData := make([]utils.CmdTuple, 0)

	messages := make([]string, 0)

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch *flag {
		case FLAG_HELP:
			return help(), nil
		case FLAG_LIST:
			tuple, err := resolveCursor(*flag, cmd)
			if err != nil {
				return "", err
			}

			return list(tuple), nil
		case FLAG_PUSH:
			tuple, err := resolveCursor(*flag, cmd)
			if err != nil {
				return "", err
			}

			pushData = append(pushData, *tuple)
		default:
			return fmt.Sprintf("Unrecognized command flag: %s", *flag), nil
		}
	}

	if len(pushData) > 0 {
		publish(user, pushData...)
	}

	return strings.Join(messages, ", "), nil
}

func help() string {
	title := "Available log actions:\n"
	return apps.RunHelp(title, refs)
}

func list(tuple *utils.CmdTuple) string {
	records := collection.VectorFromList(log.Records())

	if tuple != nil {
		switch tuple.Flag {
		case "category":
			records.FilterSelf(func(r log.Record) bool {
				return string(r.Category) == tuple.Data
			})
		}
	}

	formatter := log.Formatter{}

	return collection.VectorMap(records,
		func(r log.Record) string {
			return formatter.Format(r)
		}).Join("\n")
}

func resolveCursor(flag string, cmd *collection.Vector[string]) (*utils.CmdTuple, error) {
	value, ok := cmd.Shift()
	if !ok {
		return nil, nil
	}

	cat, mes, ok := strings.Cut(*value, "=")
	if !ok {
		return nil, fmt.Errorf("invalid flag %q value %q", flag, *value)
	}

	return utils.NewCmdTuple(cat, mes), nil
}

func publish(user string, data ...utils.CmdTuple) {
	for _, l := range data {
		message := fmt.Sprintf("(%s) - %s", user, l.Data)
		log.Custom(l.Flag, message)
	}
}
