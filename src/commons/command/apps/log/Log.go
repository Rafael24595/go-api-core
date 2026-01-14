package cmd_log

import (
	"fmt"

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
	Example:     fmt.Sprintf(`%s %s ${category}="${message}"`, Command, FLAG_PUSH),
}

func exec(user, _ string, cmd *collection.Vector[string]) *apps.CmdResult {
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
		case FLAG_LIST:
			tuple, err := apps.ResolveKeyValueCursor(cmd, "=", false)
			if err != nil {
				return apps.ErrorResult(err)
			}
			return list(tuple)
		case FLAG_PUSH:
			tuple, err := apps.ResolveKeyValueCursor(cmd, "=", true)
			if err != nil {
				return apps.ErrorResult(err)
			}
			return execPush(user, cmd, []utils.CmdTuple{*tuple})
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return apps.EmptyResult()
}

func help() *apps.CmdResult {
	title := fmt.Sprintf("Available %s actions:\n", Command)
	return apps.RunHelp(title, refs)
}

func list(tuple *utils.CmdTuple) *apps.CmdResult {
	records := collection.VectorFromList(log.Records())

	if tuple != nil {
		switch tuple.Flag {
		case "category":
			records.FilterSelf(func(r log.Record) bool {
				return string(r.Category) == tuple.Data
			})
		}
	}

	result := collection.VectorMap(records,
		func(r log.Record) string {
			return log.Formatter{}.Format(r)
		}).Join("\n")

	return apps.NewResult(result)
}

func execPush(user string, cmd *collection.Vector[string], pushData []utils.CmdTuple) *apps.CmdResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_PUSH:
			tuple, err := apps.ResolveKeyValueCursor(cmd, "=", true)
			if err != nil {
				return apps.ErrorResult(err)
			}

			pushData = append(pushData, *tuple)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	if len(pushData) > 0 {
		return publish(user, pushData...)
	}

	return apps.EmptyResult()
}

func publish(user string, data ...utils.CmdTuple) *apps.CmdResult {
	for _, l := range data {
		message := fmt.Sprintf("(%s) - %s", user, l.Data)
		log.Custom(l.Flag, message)
	}
	return apps.NewResultf("%d records pushed successfully, check the logs to see the result.", len(data))
}
