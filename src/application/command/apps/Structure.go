package apps

import (
	"fmt"

	"github.com/Rafael24595/go-api-core/src/commons/system"
	"github.com/Rafael24595/go-collections/collection"
)

type SnapshotFlag string

type CmdExecRequest struct {
	User    string
	Input   string
	Command *collection.Vector[string]
}

type CmdExecResult struct {
	Input  string
	Output string
}

func EmptyResult() *CmdExecResult {
	return &CmdExecResult{
		Input:  "",
		Output: "",
	}
}

func NewResult(output string) *CmdExecResult {
	return &CmdExecResult{
		Input:  "",
		Output: output,
	}
}

func NewResultf(format string, a ...any) *CmdExecResult {
	output := fmt.Sprintf(format, a...)
	return &CmdExecResult{
		Input:  "",
		Output: output,
	}
}

func ErrorResult(err error) *CmdExecResult {
	return &CmdExecResult{
		Input:  "",
		Output: err.Error(),
	}
}

func OverrideResult(input, output string) *CmdExecResult {
	return &CmdExecResult{
		Input:  input,
		Output: output,
	}
}

func (r *CmdExecResult) SetInput(input string) *CmdExecResult {
	r.Input = input
	return r
}

type CommandReference struct {
	Flag        SnapshotFlag
	Name        string
	Description string
	Example     string
}

type CommandApplication struct {
	CommandReference
	Exec func(request *CmdExecRequest) *CmdExecResult
	Help func() *CmdExecResult
}

type TopicMetaExpanded struct {
	Code        string
	Status      bool
	Timestamp   int64
	Description string
}

func ExpandTopicMeta(topic *system.TopicMeta, description string) TopicMetaExpanded {
	return TopicMetaExpanded{
		Code:        topic.Code,
		Status:      topic.Status,
		Timestamp:   topic.Timestamp,
		Description: description,
	}
}
