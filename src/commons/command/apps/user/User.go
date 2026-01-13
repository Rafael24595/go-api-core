package cmd_user

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/commons/session"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

const Command apps.SnapshotFlag = "user"

const (
	FLAG_HELP = "-h"
	FLAG_LIST = "-l"
)

var App = apps.CommandApplication{
	CommandReference: apps.CommandReference{
		Flag:        Command,
		Name:        "User",
		Description: "Manages system users",
		Example:     refHelp.Example,
	},
	Exec: exec,
	Help: help,
}

var refs = []apps.CommandReference{
	refHelp,
	refList,
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
	Description: "Displays the list of users.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_LIST),
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
	title := fmt.Sprintf("Available %s actions:\n", Command)
	return apps.RunHelp(title, refs)
}

func list(tuple *utils.CmdTuple) string {
	sess := repository.InstanceManagerSession()
	users := collection.VectorFromList(sess.FindAll())

	if tuple != nil {
		switch tuple.Flag {
		case "name":
			users.FilterSelf(func(u session.SessionLite) bool {
				return u.Username == tuple.Data
			})
		}
	}

	maxlen := 0
	for _, v := range users.Collect() {
		if len(v.Username) > maxlen {
			maxlen = len(v.Username)
		}
	}

	users.Sort(func(i, j session.SessionLite) bool {
		return i.Timestamp < j.Timestamp 
	})

	return collection.VectorMap(users,
		func(s session.SessionLite) string {
			space := strings.Repeat(" ", maxlen - len(s.Username))
			return fmt.Sprintf(" %s%s   %s", s.Username, space, utils.FormatMilliseconds(s.Timestamp))
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
