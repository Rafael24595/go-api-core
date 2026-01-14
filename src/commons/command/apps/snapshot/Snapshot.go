package cmd_snapshot

import (
	"fmt"
	"os"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

const Command apps.SnapshotFlag = "snpsh"

const (
	FLAG_HELP      = "-h"
	FLAG_VERBOSE   = "-v"
	FLAG_SAVE      = "-s"
	FLAG_APPLY     = "-a"
	FLAG_REMOVE    = "-rm"
	FLAG_LISTENERS = "-l"
	FLAG_DETAILS   = "-d"
)

var App = apps.CommandApplication{
	CommandReference: apps.CommandReference{
		Flag:        Command,
		Name:        "Snapshot",
		Description: "Manages in-memory persistence snapshots",
		Example:     "snpsh -h",
	},
	Exec: exec,
	Help: help,
}

var refs = []apps.CommandReference{
	refHelp,
	refListeners,
	refDetails,
	refSave,
	refApply,
	refRemove,
}

var refHelp = apps.CommandReference{
	Flag:        FLAG_HELP,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_HELP),
}

var refListeners = apps.CommandReference{
	Flag:        FLAG_LISTENERS,
	Name:        "Listeners",
	Description: "Lists all available snapshot listeners, use the verbose flag to expand the details.",
	Example:     fmt.Sprintf(`%s %s %s`, Command, FLAG_VERBOSE, FLAG_LISTENERS),
}

var refDetails = apps.CommandReference{
	Flag:        FLAG_DETAILS,
	Name:        "Details",
	Description: "Displays the list of snapshots data for a given target and type.",
	Example:     fmt.Sprintf(`%s %s snpsh_${target}=${type}`, Command, FLAG_DETAILS),
}

var refSave = apps.CommandReference{
	Flag:        FLAG_SAVE,
	Name:        "Save",
	Description: "Saves the current data as snapshot for a given target and type, use all as topic to save all data.",
	Example:     fmt.Sprintf(`%s %s snpsh_${target}=${type}`, Command, FLAG_SAVE),
}

var refApply = apps.CommandReference{
	Flag:        FLAG_APPLY,
	Name:        "Apply",
	Description: "Applies a previously saved input snapshot for a given target and type retrieved from the extension.",
	Example:     fmt.Sprintf(`%s %s snpsh_${target}=${name}.${extension}`, Command, FLAG_APPLY),
}

var refRemove = apps.CommandReference{
	Flag:        FLAG_REMOVE,
	Name:        "Remove",
	Description: "Removes a previously saved snapshot for a given target and type retrieved from the extension.",
	Example:     fmt.Sprintf(`%s %s snpsh_${target}=${name}.${extension}`, Command, FLAG_REMOVE),
}

func exec(request *apps.CmdRequest) *apps.CmdResult {
	cmd := request.Command
	
	if cmd.Size() == 0 {
		return help()
	}

	verbose := false

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_HELP:
			return help()
		case FLAG_VERBOSE:
			verbose = true
		case FLAG_LISTENERS:
			return execListerner(cmd, verbose)
		case FLAG_DETAILS:
			return details(cmd)
		case FLAG_SAVE:
			tuples, err := save(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}
			return execSave(cmd, tuples)
		case FLAG_APPLY:
			topic, value, err := resolveCursor(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			tuple := utils.NewCmdTuple(topic.TopicSnapshotAppyInput(), value)
			return execApply(cmd, []utils.CmdTuple{*tuple})
		case FLAG_REMOVE:
			topic, value, err := resolveCursor(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			tuple := utils.NewCmdTuple(topic.TopicSnapshotRemoveInput(), value)
			return execRemove(cmd, []utils.CmdTuple{*tuple})
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return apps.EmptyResult()
}

func help()  *apps.CmdResult {
	title := fmt.Sprintf("Available %s actions:\n", Command)
	return apps.RunHelp(title, refs)
}

func execListerner(cmd *collection.Vector[string], verbose bool) *apps.CmdResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_VERBOSE:
			verbose = true
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return listeners(verbose)
}

func listeners(verbose bool) *apps.CmdResult {
	config := configuration.Instance()
	raw := config.EventHub.Topics(repository.SnapshotListener)
	listeners := system.FilterTopicSnapshot(raw)

	result := make([]string, len(listeners))

	for i, v := range listeners {
		var listener string
		if verbose {
			listener = fmt.Sprintf(" - %s: %s", v, v.Description())
		} else {
			listener = string(v)
		}
		result[i] = listener
	}

	if verbose {
		header := fmt.Sprintf("Active listeners [%d]", len(result))
		result = append([]string{header}, result...)
	}

	format := fmt.Sprintf("Active format [%s]", config.Format())
	result = append([]string{format, ""}, result...)

	return apps.NewResult(strings.Join(result, "\n"))
}

func details(cmd *collection.Vector[string]) *apps.CmdResult {
	path, err := findDetailsPath(cmd)
	if err != nil {
		return apps.ErrorResult(err)
	}

	files, err := repository.FindSnapshots(path)
	if err != nil {
		return apps.ErrorResult(err)
	}

	snapshots := collection.VectorMap(files, func(e os.DirEntry) string {
		return e.Name()
	})

	result := strings.Join(snapshots.Collect(), "\n")

	return apps.NewResult(result)
}

func execSave(cmd *collection.Vector[string], data []utils.CmdTuple) *apps.CmdResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_SAVE:
			tuples, err := save(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}
			data = append(data, tuples...)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	if len(data) > 0 {
		return publishEvent(data...)
	}

	return apps.EmptyResult()
}

func execApply(cmd *collection.Vector[string], data []utils.CmdTuple) *apps.CmdResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_APPLY:
			topic, value, err := resolveCursor(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			tuple := utils.NewCmdTuple(topic.TopicSnapshotAppyInput(), value)
			data = append(data, *tuple)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	if len(data) > 0 {
		return publishEvent(data...)
	}

	return apps.EmptyResult()
}

func execRemove(cmd *collection.Vector[string], data []utils.CmdTuple) *apps.CmdResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_REMOVE:
			topic, value, err := resolveCursor(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			tuple := utils.NewCmdTuple(topic.TopicSnapshotRemoveInput(), value)
			data = append(data, *tuple)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	if len(data) > 0 {
		return publishEvent(data...)
	}

	return apps.EmptyResult()
}

func save(cmd *collection.Vector[string]) ([]utils.CmdTuple, error) {
	cmds := make([]utils.CmdTuple, 0)

	tuple, err := apps.ResolveKeyValueCursor(cmd, "=", true)
	if err != nil {
		return cmds, err
	}

	tpc := tuple.Flag
	snpsh := tuple.Data

	var topics []system.TopicSnapshot
	if tpc == "all" {
		config := configuration.Instance()
		raw := config.EventHub.Topics(repository.SnapshotListener)
		topics = system.FilterTopicSnapshot(raw)
	} else {
		topic, ok := system.TopicSnapshotFromString(tpc)
		if !ok {
			return cmds, fmt.Errorf("unknown topic: %s", tpc)
		}
		topics = []system.TopicSnapshot{topic}
	}

	for _, v := range topics {
		tuple := utils.NewCmdTuple(v.TopicSnapshotSaveInput(), snpsh)
		cmds = append(cmds, *tuple)
	}

	return cmds, nil
}

func publishEvent(data ...utils.CmdTuple) *apps.CmdResult {
	config := configuration.Instance()
	for _, cmd := range data {
		config.EventHub.Publish(cmd.Flag, cmd.Data)
	}
	return apps.NewResultf("%d events pushed successfully, check the logs to see the result.", len(data))
}

func findDetailsPath(cmd *collection.Vector[string]) (string, error) {
	tuple, err := apps.ResolveKeyValueCursor(cmd, "=", true)
	if err != nil {
		return "", err
	}

	tpc := tuple.Flag
	ext := tuple.Data

	topic, ok := system.TopicSnapshotFromString(tpc)
	if !ok {
		return "", fmt.Errorf("unknown topic: %s", tpc)
	}

	format, ok := format.DataFormatFromExtension(ext)
	if !ok {
		return "", fmt.Errorf("undefined format for extension %q", ext)
	}

	path, ok := topic.Path(format)
	if !ok {
		return "", fmt.Errorf("undefined path for extension %q", ext)
	}

	return path, nil
}

func resolveCursor(cmd *collection.Vector[string]) (system.TopicSnapshot, string, error) {
	tuple, err := apps.ResolveKeyValueCursor(cmd, "=", true)
	if err != nil {
		return "", "", err
	}

	tpc := tuple.Flag
	snpsh := tuple.Data

	topic, ok := system.TopicSnapshotFromString(tpc)
	if !ok {
		return "", "", fmt.Errorf("unknown topic: %s", tpc)
	}

	return topic, snpsh, nil
}
