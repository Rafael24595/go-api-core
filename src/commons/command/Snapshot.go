package command

import (
	"fmt"
	"os"
	"strings"

	command_helper "github.com/Rafael24595/go-api-core/src/commons/command/helper"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

const (
	FLAG_SNAPSHOT_HELP      = "-h"
	FLAG_SNAPSHOT_VERBOSE   = "-v"
	FLAG_SNAPSHOT_SAVE      = "-s"
	FLAG_SNAPSHOT_APPLY     = "-a"
	FLAG_SNAPSHOT_REMOVE    = "-r"
	FLAG_SNAPSHOT_LISTENERS = "-l"
	FLAG_SNAPSHOT_DETAILS   = "-d"
)

var snapshotHelp = commandReference{
	Flag:        FLAG_SNAPSHOT_HELP,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     `snpsh -h`,
}

var snapshotSave = commandReference{
	Flag:        FLAG_SNAPSHOT_SAVE,
	Name:        "Save",
	Description: "Saves the current data as snapshot for a given target and type, use all as topic to save all data.",
	Example:     `snpsh -s snpsh_${target}=${type}`,
}

var snapshotApply = commandReference{
	Flag:        FLAG_SNAPSHOT_APPLY,
	Name:        "Apply",
	Description: "Applies a previously saved input snapshot for a given target and type retrieved from the extension.",
	Example:     `snpsh -a snpsh_${target}=${name}.${extension}`,
}

var snapshotRemove = commandReference{
	Flag:        FLAG_SNAPSHOT_REMOVE,
	Name:        "Remove",
	Description: "Removes a previously saved snapshot for a given target and type retrieved from the extension.",
	Example:     `snpsh -r snpsh_${target}=${name}.${extension}`,
}

var snapshotDetails = commandReference{
	Flag:        FLAG_SNAPSHOT_DETAILS,
	Name:        "Details",
	Description: "Displays the list of snapshots data for a given target and type.",
	Example:     `snpsh -d snpsh_${target}=${type}`,
}

var snapshotListeners = commandReference{
	Flag:        FLAG_SNAPSHOT_LISTENERS,
	Name:        "Listeners",
	Description: "Lists all available snapshot listeners, use the verbose flag to expand the details.",
	Example:     `snpsh -v -l`,
}

var snapshotActions = []commandReference{
	snapshotHelp,
	snapshotSave,
	snapshotApply,
	snapshotRemove,
	snapshotDetails,
	snapshotListeners,
}

func snapshot(_ string, cmd *collection.Vector[string]) (string, error) {
	if cmd.Size() == 0 {
		return runSnapshotHelp(), nil
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
		case FLAG_SNAPSHOT_HELP:
			return runSnapshotHelp(), nil
		case FLAG_SNAPSHOT_VERBOSE:
			verbose = true
		case FLAG_SNAPSHOT_SAVE:
			tuples, err := runSnapshotSave(*flag, cmd)
			if err != nil {
				return "", err
			}

			saveData = append(saveData, tuples...)

		case FLAG_SNAPSHOT_APPLY:
			topic, value, err := runSnapshotCursor(*flag, cmd)
			if err != nil {
				return "", err
			}

			tuple := utils.NewCmdTuple(topic.TopicSnapshotAppyInput(), value)
			applyData = append(applyData, *tuple)

		case FLAG_SNAPSHOT_REMOVE:
			topic, value, err := runSnapshotCursor(*flag, cmd)
			if err != nil {
				return "", err
			}

			tuple := utils.NewCmdTuple(topic.TopicSnapshotRemoveInput(), value)
			removeData = append(removeData, *tuple)

		case FLAG_SNAPSHOT_LISTENERS:
			message := runSnapshotListeners(verbose)
			messages = append(messages, message)

		case FLAG_SNAPSHOT_DETAILS:
			message, err := runSnapshotDetails(*flag, cmd)
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

func runSnapshotHelp() string {
	title := "Available snapshot actions:\n"
	return runHelp(title, snapshotActions)
}

func runSnapshotCursor(flag string, cmd *collection.Vector[string]) (system.TopicSnapshot, string, error) {
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

func runSnapshotSave(flag string, cmd *collection.Vector[string]) ([]utils.CmdTuple, error) {
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

func runSnapshotListeners(verbose bool) string {
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

func runSnapshotDetails(flag string, cmd *collection.Vector[string]) (string, error) {
	path, err := findSnapshotDetailsPath(flag, cmd)
	if err != nil {
		return "", err
	}

	files, err := command_helper.FindSnapshots(path)
	if err != nil {
		return "", err
	}

	snapshots := collection.VectorMap(files, func(e os.DirEntry) string {
		return e.Name()
	})

	return strings.Join(snapshots.Collect(), "\n"), nil
}

func findSnapshotDetailsPath(flag string, cmd *collection.Vector[string]) (string, error) {
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
