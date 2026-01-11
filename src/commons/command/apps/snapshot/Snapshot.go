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
	FLAG_REMOVE    = "-r"
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
	refSave,
	refApply,
	refRemove,
	refDetails,
	refListeners,
}

var refHelp = apps.CommandReference{
	Flag:        FLAG_HELP,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_HELP),
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

var refDetails = apps.CommandReference{
	Flag:        FLAG_DETAILS,
	Name:        "Details",
	Description: "Displays the list of snapshots data for a given target and type.",
	Example:     fmt.Sprintf(`%s %s snpsh_${target}=${type}`, Command, FLAG_DETAILS),
}

var refListeners = apps.CommandReference{
	Flag:        FLAG_LISTENERS,
	Name:        "Listeners",
	Description: "Lists all available snapshot listeners, use the verbose flag to expand the details.",
	Example:     fmt.Sprintf(`%s %s %s`, Command, FLAG_VERBOSE, FLAG_LISTENERS),
}

func exec(_ string, cmd *collection.Vector[string]) (string, error) {
	if cmd.Size() == 0 {
		return help(), nil
	}

	verbose := false

	saveData := make([]utils.CmdTuple, 0)
	applyData := make([]utils.CmdTuple, 0)
	removeData := make([]utils.CmdTuple, 0)

	messages := make([]string, 0)

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch *flag {
		case FLAG_HELP:
			return help(), nil
		case FLAG_VERBOSE:
			verbose = true
		case FLAG_SAVE:
			tuples, err := save(*flag, cmd)
			if err != nil {
				return "", err
			}

			saveData = append(saveData, tuples...)

		case FLAG_APPLY:
			topic, value, err := resolveCursor(*flag, cmd)
			if err != nil {
				return "", err
			}

			tuple := utils.NewCmdTuple(topic.TopicSnapshotAppyInput(), value)
			applyData = append(applyData, *tuple)

		case FLAG_REMOVE:
			topic, value, err := resolveCursor(*flag, cmd)
			if err != nil {
				return "", err
			}

			tuple := utils.NewCmdTuple(topic.TopicSnapshotRemoveInput(), value)
			removeData = append(removeData, *tuple)

		case FLAG_LISTENERS:
			message := listeners(verbose)
			messages = append(messages, message)

		case FLAG_DETAILS:
			message, err := details(*flag, cmd)
			if err != nil {
				return "", err
			}

			messages = append(messages, message)
		default:
			return fmt.Sprintf("Unrecognized command flag: %s", *flag), nil
		}
	}

	if len(saveData) > 0 {
		messages = append(messages, publishEvent(saveData...))
	}

	if len(applyData) > 0 {
		messages = append(messages, publishEvent(applyData...))
	}

	if len(removeData) > 0 {
		messages = append(messages, publishEvent(removeData...))
	}

	return strings.Join(messages, ", "), nil
}

func help() string {
	title := "Available snapshot actions:\n"
	return apps.RunHelp(title, refs)
}

func save(flag string, cmd *collection.Vector[string]) ([]utils.CmdTuple, error) {
	cmds := make([]utils.CmdTuple, 0)

	value, ok := cmd.Shift()
	if !ok {
		return cmds, nil
	}

	tpc, snpsh, ok := strings.Cut(*value, "=")
	if !ok {
		return cmds, fmt.Errorf("invalid flag %q value %q", flag, *value)
	}

	var topics []system.TopicSnapshot
	if tpc == "all" {
		config := configuration.Instance()
		raw := config.EventHub.Topics(repository.SnapshotListener)
		topics = system.FilterTopicSnapshot(raw)
	} else {
		topic, ok := system.TopicSnapshotFromString(tpc)
		if !ok {
			return cmds, fmt.Errorf("unknown topic: %s", *value)
		}
		topics = []system.TopicSnapshot{topic}
	}

	for _, v := range topics {
		tuple := utils.NewCmdTuple(v.TopicSnapshotSaveInput(), snpsh)
		cmds = append(cmds, *tuple)
	}

	return cmds, nil
}

func publishEvent(data ...utils.CmdTuple) string {
	config := configuration.Instance()
	for _, cmd := range data {
		config.EventHub.Publish(cmd.Flag, cmd.Data)
	}
	return fmt.Sprintf("%d events pushed successfully, check the logs to see the result.", len(data))
}

func listeners(verbose bool) string {
	config := configuration.Instance()
	raw := config.EventHub.Topics(repository.SnapshotListener)
	listeners := system.FilterTopicSnapshot(raw)

	result := make([]string, len(listeners))

	for i, v := range listeners {
		var listener string
		if verbose {
			listener = fmt.Sprintf(" - %s: %s", listener, v.Description())
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

	return strings.Join(result, "\n")
}

func details(flag string, cmd *collection.Vector[string]) (string, error) {
	path, err := findDetailsPath(flag, cmd)
	if err != nil {
		return "", err
	}

	files, err := repository.FindSnapshots(path)
	if err != nil {
		return "", err
	}

	snapshots := collection.VectorMap(files, func(e os.DirEntry) string {
		return e.Name()
	})

	return strings.Join(snapshots.Collect(), "\n"), nil
}

func findDetailsPath(flag string, cmd *collection.Vector[string]) (string, error) {
	value, ok := cmd.Shift()
	if !ok {
		return "", nil
	}

	tpc, ext, ok := strings.Cut(*value, "=")
	if !ok {
		return "", fmt.Errorf("invalid flag %q value %q", flag, *value)
	}

	topic, ok := system.TopicSnapshotFromString(tpc)
	if !ok {
		return "", fmt.Errorf("unknown topic: %s", *value)
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

func resolveCursor(flag string, cmd *collection.Vector[string]) (system.TopicSnapshot, string, error) {
	value, ok := cmd.Shift()
	if !ok {
		return "", "", nil
	}

	tpc, snpsh, ok := strings.Cut(*value, "=")
	if !ok {
		return "", "", fmt.Errorf("invalid flag %q value %q", flag, *value)
	}

	topic, ok := system.TopicSnapshotFromString(tpc)
	if !ok {
		return "", "", fmt.Errorf("unknown topic: %s", *value)
	}

	return topic, snpsh, nil
}
