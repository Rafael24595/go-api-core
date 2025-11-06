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
	FLAG_SNAPSHOT_VERBOSE   = "-v"
	FLAG_SNAPSHOT_SAVE      = "-s"
	FLAG_SNAPSHOT_APPLY     = "-a"
	FLAG_SNAPSHOT_LISTENERS = "-l"
	FLAG_SNAPSHOT_DETAILS   = "-d"
)

func snapshot(cmd *collection.Vector[string]) (string, error) {
	verbose := false
	saveData := make([]utils.CmdTuple, 0)
	applyData := make([]utils.CmdTuple, 0)

	messages := make([]string, 0)

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch *flag {
		case FLAG_SNAPSHOT_VERBOSE:
			verbose = true
		case FLAG_SNAPSHOT_SAVE:
			data := runSnapshotSave(cmd)
			if data != nil {
				applyData = append(saveData, *data)
			}

		case FLAG_SNAPSHOT_APPLY:
			data, err := runSnapshotApply(*flag, cmd)
			if err != nil {
				return "", err
			}

			if data != nil {
				applyData = append(applyData, *data)
			}

		case FLAG_SNAPSHOT_LISTENERS:
			message := runSnapshotListeners(verbose)
			messages = append(messages, message)

		case FLAG_SNAPSHOT_DETAILS:
			message, err := runSnapshotDetails(*flag, cmd)
			if err != nil {
				return "", err
			}

			messages = append(messages, message)
		}
	}

	if len(saveData) > 0 {
		message, err := snapshotSave(saveData)
		if err != nil {
			return "", err
		}
		messages = append(messages, message)
	}

	if len(applyData) > 0 {
		message, err := snapshotApply(applyData)
		if err != nil {
			return "", err
		}
		messages = append(messages, message)
	}

	return strings.Join(messages, ", "), nil
}

func runSnapshotSave(cmd *collection.Vector[string]) *utils.CmdTuple {
	value, ok := cmd.Shift()
	if !ok {
		return nil
	}

	topic, snpsh, ok := strings.Cut(*value, "=")
	if ok {
		return utils.NewCmdTuple(topic, snpsh)
	}

	return utils.NewCmdTuple(*value, "")
}

func snapshotSave(data []utils.CmdTuple) (string, error) {
	config := configuration.Instance()
	for _, cmd := range data {
		config.EventHub.Publish(cmd.Flag, cmd.Data)
	}
	return "", nil
}

func runSnapshotApply(flag string, cmd *collection.Vector[string]) (*utils.CmdTuple, error) {
	value, ok := cmd.Shift()
	if !ok {
		return nil, nil
	}

	topic, snpsh, ok := strings.Cut(*value, "=")
	if !ok {
		return nil, fmt.Errorf("invalid flag %q value %q", flag, *value)
	}

	return utils.NewCmdTuple(topic, snpsh), nil
}

func snapshotApply(data []utils.CmdTuple) (string, error) {
	config := configuration.Instance()
	for _, cmd := range data {
		config.EventHub.Publish(cmd.Flag, cmd.Data)
	}
	return "", nil
}

func runSnapshotListeners(verbose bool) string {
	config := configuration.Instance()
	raw := config.EventHub.Topics(repository.SnapshotListener)
	listeners := system.FilterTopicSnapshot(raw)

	result := make([]string, len(listeners))

	for i, v := range listeners {
		listener := string(v)
		if verbose {
			listener = fmt.Sprintf(" - %s: %s", listener, v.Description())
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
