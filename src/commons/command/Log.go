package command

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-collections/collection"
)

const (
	FLAG_LOG_HELP = "-h"
	FLAG_LOG_LIST = "-l"
	FLAG_LOG_PUSH = "-p"
)

var logHelp = commandAction{
	Flag:        FLAG_LOG_HELP,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     `log -h`,
}

var logList = commandAction{
	Flag:        FLAG_LOG_LIST,
	Name:        "List",
	Description: "Displays the list of log records.",
	Example:     `log -l`,
}

var logPush = commandAction{
	Flag:        FLAG_LOG_PUSH,
	Name:        "Push",
	Description: "Insert a new log record.",
	Example:     `log -p ${category}=${message}`,
}

var logActions = []commandAction{
	logHelp,
	logList,
	logPush,
}

func logg(cmd *collection.Vector[string]) (string, error) {
	if cmd.Size() == 0 {
		return runLogHelp(), nil
	}

	pushData := make([]utils.CmdTuple, 0)

	messages := make([]string, 0)

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch *flag {
		case FLAG_LOG_HELP:
			return runLogHelp(), nil
		case FLAG_LOG_LIST:
			tuple, err := runLogCursor(*flag, cmd)
			if err != nil {
				return "", err
			}

			return runLogList(tuple), nil
		case FLAG_LOG_PUSH:
			tuple, err := runLogCursor(*flag, cmd)
			if err != nil {
				return "", err
			}

			pushData = append(pushData, *tuple)
		}
	}

	if len(pushData) > 0 {
		publishLog(pushData...)
	}

	return strings.Join(messages, ", "), nil
}

func runLogHelp() string {
	title := "Available log actions:\n"
	return runHelp(title, logActions)
}

func runLogList(tuple *utils.CmdTuple) string {
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

func runLogCursor(flag string, cmd *collection.Vector[string]) (*utils.CmdTuple, error) {
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

func publishLog(data ...utils.CmdTuple) {
	for _, l := range data {
		log.Custom(l.Flag, l.Data)
	}
}
