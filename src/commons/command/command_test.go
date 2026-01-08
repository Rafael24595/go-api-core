package command

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/commons/command/apps"
	"github.com/Rafael24595/go-api-core/test/support/assert"
)

var testCmdActions = []apps.CommandApplication{
	{
		CommandReference: apps.CommandReference{
			Flag: "snapshot",
		},
		Help: func() string { return "Test help message" },
	},
	{
		CommandReference: apps.CommandReference{
			Flag: "system",
		},
	},
}

func TestComplete_WhenSingleMatch_ReturnsCorrectAction(t *testing.T) {
	cmd := "sna"
	coincidences, cursor, position := comp(cmd, InitialStep, testCmdActions)

	action := &testCmdActions[0]

	assert.NotNil(t, cursor)

	assert.Equal(t, position, 0)
	assert.Equal(t, action.Flag, cursor.Flag)
	assert.Equal(t, 1, len(coincidences))
	assert.Equal(t, cursor.Help(), action.Help())
}

func TestComplete_WhenMultipleMatches_CyclesThroughActions(t *testing.T) {
	cmd := "s"
	lastPosition := InitialStep

	for i := range 6 {
		coincidences, cursor, position := comp(cmd, lastPosition, testCmdActions)
		lastPosition = position

		focus := i
		if i >= len(testCmdActions) {
			focus = i % len(testCmdActions)
		}

		action := &testCmdActions[focus]

		assert.NotNil(t, cursor)

		assert.Equal(t, position, focus)
		assert.Equal(t, action.Flag, cursor.Flag)
		assert.Equal(t, 2, len(coincidences))
	}
}

func TestComplete_WhenPartialCommandMatches_ReturnsHelpContext(t *testing.T) {
	cmd := "snp"
	help, err := Comp("anonymous", cmd, 999)

	assert.NotError(t, err)
	assert.NotNil(t, help)

	action := &refApps[2]

	assert.Equal(t, help.Position, 2)
	assert.Equal(t, string(action.Flag), help.Application)
	assert.Equal(t, "", help.Message)
}

func TestComplete_WhenCommandDoesNotExist_ReturnsEmptyHelp(t *testing.T) {
	cmd := "undefined"
	help, err := Comp("anonymous", cmd, InitialStep)

	assert.NotError(t, err)
	assert.NotNil(t, help)

	assert.Equal(t, help.Position, InitialStep)
	assert.Equal(t, "", help.Application)
	assert.Equal(t, "", help.Message)
}

func TestComplete_WhenExactCommandProvided_ReturnsCommandHelp(t *testing.T) {
	action := &refApps[0]

	help, err := Comp("anonymous", string(action.Flag), InitialStep)

	assert.NotError(t, err)
	assert.NotNil(t, help)

	assert.Equal(t, action.Help(), help.Message)
}
