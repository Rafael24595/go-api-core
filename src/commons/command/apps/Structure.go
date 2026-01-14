package apps

import (
	"fmt"

	"github.com/Rafael24595/go-collections/collection"
)

type SnapshotFlag string

type CmdResult struct {
	Input  string
	Output string
}

func EmptyResult() *CmdResult {
	return &CmdResult{
		Input:  "",
		Output: "",
	}
}

func NewResult(output string) *CmdResult {
	return &CmdResult{
		Input:  "",
		Output: output,
	}
}

func NewResultf(format string, a ...any) *CmdResult {
	output := fmt.Sprintf(format, a...)
	return &CmdResult{
		Input:  "",
		Output: output,
	}
}

func ErrorResult(err error) *CmdResult {
	return &CmdResult{
		Input:  "",
		Output: err.Error(),
	}
}

func OverrideResult(input, output string) *CmdResult {
	return &CmdResult{
		Input:  input,
		Output: output,
	}
}

func (r *CmdResult) SetInput(input string) *CmdResult {
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
	Exec func(user, request string, cmd *collection.Vector[string]) *CmdResult
	Help func() *CmdResult
}
