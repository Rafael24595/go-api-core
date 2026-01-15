package cmd_repo

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	"github.com/Rafael24595/go-api-core/src/commons/configuration"
	"github.com/Rafael24595/go-api-core/src/commons/system"
	topic_repository "github.com/Rafael24595/go-api-core/src/commons/system/topic/repository"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/infrastructure/repository"
	"github.com/Rafael24595/go-collections/collection"
)

const Command apps.SnapshotFlag = "repo"

const (
	FLAG_HELP      = "-h"
	FLAG_VERBOSE   = "-v"
	FLAG_LISTENERS = "-l"
	FLAG_DETAILS   = "-d"
	FLAG_RELOAD    = "-r"
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
	Description: "Lists all available resolver listeners, use the verbose flag to expand the details.",
	Example:     fmt.Sprintf(`%s %s %s`, Command, FLAG_VERBOSE, FLAG_LISTENERS),
}

var refDetails = apps.CommandReference{
	Flag:        FLAG_DETAILS,
	Name:        "Details",
	Description: "Displays the list of resolver listeners with it's status and metadata for a given targets.",
	Example:     fmt.Sprintf(`%s %s ${topic}+${topic}`, Command, FLAG_DETAILS),
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
		case FLAG_RELOAD:
			topics, err := resolveTopics(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			return execReload(cmd, topics)
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
	topics := config.EventHub.TopicsMeta(repository.RepositoryListener)
	parents := findParents(topics)

	if verbose {
		result := formatParentsVerbose(parents...)
		return apps.NewResult(result)
	}

	result := formatParents(parents...)
	return apps.NewResult(result)
}

func execDetails(cmd *collection.Vector[string], topics []topic_repository.TopicRepository) *apps.CmdExecResult {
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

func details(filter []topic_repository.TopicRepository) *apps.CmdExecResult {
	config := configuration.Instance()
	topics := config.EventHub.TopicsMeta(repository.RepositoryListener)

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

func formatParents(topics ...topic_repository.TopicRepository) string {
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

func formatParentsVerbose(topics ...topic_repository.TopicRepository) string {
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

		topic, ok := topic_repository.TopicFromString(v.Parent)
		if ok {
			if action, ok := topic.FindAcction(v.Code); ok {
				description = action.Description
			}
		}

		buffer[i] = apps.ExpandTopicMeta(&v, description)
	}

	return buffer
}

func execReload(cmd *collection.Vector[string], topics []topic_repository.TopicRepository) *apps.CmdExecResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_RELOAD:
			result, err := resolveTopics(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}

			topics = append(topics, result...)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	tuples := make([]utils.CmdTuple, len(topics))
	for i, v := range topics {
		tuples[i] = utils.CmdTuple{
			Flag: v.ActionReload().Code,
		}
	}

	if len(topics) > 0 {
		return apps.PublishEvent(tuples...)
	}

	return apps.EmptyResult()
}

func resolveTopics(cmd *collection.Vector[string]) ([]topic_repository.TopicRepository, error) {
	raw, err := apps.ResolveChainCursor(cmd, "+")
	if err != nil {
		return make([]topic_repository.TopicRepository, 0), err
	}

	topics := make([]topic_repository.TopicRepository, len(raw))
	for i, v := range raw {
		topic, ok := topic_repository.TopicFromString(v)
		if !ok {
			return topics, fmt.Errorf("unknown topic: %s", v)
		}
		topics[i] = topic
	}

	return topics, nil
}

func findParents(topics []system.TopicMeta) []topic_repository.TopicRepository {
	cache := make(map[string]topic_repository.TopicRepository)

	for _, v := range topics {
		if _, ok := cache[v.Parent]; ok {
			continue
		}

		topic, ok := topic_repository.TopicFromString(v.Parent)
		if ok {
			cache[v.Parent] = topic
		}
	}

	return collection.DictionaryFromMap(cache).Values()
}
