package cmd_snapshot

import (
	"fmt"
	"os"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/format"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	topic_snapshot "github.com/Rafael24595/go-api-core/src/commons/system/topic/snapshot"
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
	FLAG_FILES     = "-f"
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
	refFiles,
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
	Description: "Displays the list of snapshot listeners with it's status and metadata for a given targets.",
	Example:     fmt.Sprintf(`%s %s ${topic}+${topic}`, Command, FLAG_DETAILS),
}

var refFiles = apps.CommandReference{
	Flag:        FLAG_FILES,
	Name:        "Details",
	Description: "Displays the list of snapshots files for a given target and type (e.g., csvt).",
	Example:     fmt.Sprintf(`%s %s ${topic}=${type}`, Command, FLAG_DETAILS),
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

func exec(request *apps.CmdExecRequest) *apps.CmdExecResult {
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
			filter, err := resolveTopics(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			return execDetails(cmd, filter)
		case FLAG_FILES:
			path, err := findDetailsPath(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			return execFiles(cmd, path)
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

			tuple := utils.NewCmdTuple(topic.ActionAppy().Code, value)
			return execApply(cmd, []utils.CmdTuple{*tuple})
		case FLAG_REMOVE:
			topic, value, err := resolveCursor(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			tuple := utils.NewCmdTuple(topic.ActionRemove().Code, value)
			return execRemove(cmd, []utils.CmdTuple{*tuple})
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

func execListerner(cmd *collection.Vector[string], verbose bool) *apps.CmdExecResult {
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

func listeners(verbose bool) *apps.CmdExecResult {
	config := configuration.Instance()
	topics := config.EventHub.TopicsMeta(repository.SnapshotListener)
	parents := findParents(topics)

	if verbose {
		result := formatParentsVerbose(parents...)
		return apps.NewResult(result)
	}

	result := formatParents(parents...)
	return apps.NewResult(result)
}

func execDetails(cmd *collection.Vector[string], topics []topic_snapshot.TopicSnapshot) *apps.CmdExecResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_DETAILS:
			result, err := resolveTopics(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}
			topics = append(topics, result...)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return details(topics)
}

func details(filter []topic_snapshot.TopicSnapshot) *apps.CmdExecResult {
	config := configuration.Instance()
	topics := config.EventHub.TopicsMeta(repository.SnapshotListener)

	result := make([]system.TopicMeta, 0)

	for _, f := range filter {
		for _, e := range topics {
			if e.Parent == string(f) {
				result = append(result, e)
			}
		}
	}

	extend := extendTopicMeta(result)

	format := formatTopicsVerbose(extend...)
	return apps.NewResult(format)
}

func execFiles(cmd *collection.Vector[string], paths ...string) *apps.CmdExecResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_FILES:
			path, err := findDetailsPath(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			paths = append(paths, path)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return files(paths)
}

func files(paths []string) *apps.CmdExecResult {
	buffer := collection.VectorEmpty[os.DirEntry]()

	for _, v := range paths {
		files, err := repository.FindSnapshots(v)
		if err != nil {
			return apps.ErrorResult(err)
		}
	
		buffer.Merge(*files)
	}

	snapshots := collection.VectorMap(buffer, func(e os.DirEntry) string {
		return e.Name()
	})

	result := strings.Join(snapshots.Collect(), "\n")

	return apps.NewResult(result)
}

func formatParents(topics ...topic_snapshot.TopicSnapshot) string {
	maxlen := 0

	for _, v := range topics {
		maxlen = utils.Max(maxlen, len(v))
	}

	buffer := make([]string, len(topics))

	for i, v := range topics {
		fcode := utils.Right(v, maxlen)
		buffer[i] = fmt.Sprintf(" - %s   %s", fcode, v.Description())
	}

	return strings.Join(buffer, "\n")
}

func formatParentsVerbose(topics ...topic_snapshot.TopicSnapshot) string {
	table := utils.NewTable()

	table.Headers("Code", "Description")

	for i, v := range topics {
		table.Field("Code", i, v)
		table.Field("Description", i, v.Description())
	}

	return table.ToString()
}

func formatTopicsVerbose(topics ...apps.TopicMetaExpanded) string {
	table := utils.NewTable()

	table.Headers("Code", "Status", "Timestamp", "Description")

	for i, v := range topics {
		table.Field("Code", i, v.Code)
		table.Field("Status", i, v.Status)
		table.Field("Timestamp", i, utils.FormatMilliseconds(v.Timestamp))
		table.Field("Description", i, v.Description)
	}

	return table.ToString()
}

func extendTopicMeta(topics []system.TopicMeta) []apps.TopicMetaExpanded {
	buffer := make([]apps.TopicMetaExpanded, len(topics))

	for i, v := range topics {
		description := ""

		topic, ok := topic_snapshot.TopicFromString(v.Parent)
		if ok {
			if action, ok := topic.FindAcction(v.Code); ok {
				description = action.Description
			}
		}

		buffer[i] = apps.ExpandTopicMeta(&v, description)
	}

	return buffer
}

func execSave(cmd *collection.Vector[string], data []utils.CmdTuple) *apps.CmdExecResult {
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
		return apps.PublishEvent(data...)
	}

	return apps.EmptyResult()
}

func execApply(cmd *collection.Vector[string], data []utils.CmdTuple) *apps.CmdExecResult {
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

			tuple := utils.NewCmdTuple(topic.ActionAppy().Code, value)
			data = append(data, *tuple)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	if len(data) > 0 {
		return apps.PublishEvent(data...)
	}

	return apps.EmptyResult()
}

func execRemove(cmd *collection.Vector[string], data []utils.CmdTuple) *apps.CmdExecResult {
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

			tuple := utils.NewCmdTuple(topic.ActionRemove().Code, value)
			data = append(data, *tuple)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	if len(data) > 0 {
		return apps.PublishEvent(data...)
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

	var topics []topic_snapshot.TopicSnapshot
	if tpc == "all" {
		config := configuration.Instance()
		raw := config.EventHub.Topics(repository.SnapshotListener)
		topics = topic_snapshot.FindTopics(raw)
	} else {
		topic, ok := topic_snapshot.TopicFromString(tpc)
		if !ok {
			return cmds, fmt.Errorf("unknown topic: %s", tpc)
		}
		topics = []topic_snapshot.TopicSnapshot{topic}
	}

	for _, v := range topics {
		tuple := utils.NewCmdTuple(v.ActionSave().Code, snpsh)
		cmds = append(cmds, *tuple)
	}

	return cmds, nil
}

func findDetailsPath(cmd *collection.Vector[string]) (string, error) {
	tuple, err := apps.ResolveKeyValueCursor(cmd, "=", true)
	if err != nil {
		return "", err
	}

	tpc := tuple.Flag
	ext := tuple.Data

	topic, ok := topic_snapshot.TopicFromString(tpc)
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

func resolveTopics(cmd *collection.Vector[string]) ([]topic_snapshot.TopicSnapshot, error) {
	raw, err := apps.ResolveChainCursor(cmd, "+")
	if err != nil {
		return make([]topic_snapshot.TopicSnapshot, 0), err
	}

	topics := make([]topic_snapshot.TopicSnapshot, len(raw))
	for i, v := range raw {
		topic, ok := topic_snapshot.TopicFromString(v)
		if !ok {
			return topics, fmt.Errorf("unknown topic: %s", v)
		}
		topics[i] = topic
	}

	return topics, nil
}

func resolveCursor(cmd *collection.Vector[string]) (topic_snapshot.TopicSnapshot, string, error) {
	tuple, err := apps.ResolveKeyValueCursor(cmd, "=", true)
	if err != nil {
		return "", "", err
	}

	tpc := tuple.Flag
	snpsh := tuple.Data

	topic, ok := topic_snapshot.TopicFromString(tpc)
	if !ok {
		return "", "", fmt.Errorf("unknown topic: %s", tpc)
	}

	return topic, snpsh, nil
}

func findParents(topics []system.TopicMeta) []topic_snapshot.TopicSnapshot {
	cache := make(map[string]topic_snapshot.TopicSnapshot)

	for _, v := range topics {
		if _, ok := cache[v.Parent]; ok {
			continue
		}

		topic, ok := topic_snapshot.TopicFromString(v.Parent)
		if ok {
			cache[v.Parent] = topic
		}
	}

	return collection.DictionaryFromMap(cache).Values()
}
