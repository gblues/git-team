package statuscmdadapter

import (
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/hekmekk/git-team/src/command/adapter"
	"github.com/hekmekk/git-team/src/command/status"
	"github.com/hekmekk/git-team/src/command/status/interfaceadapter/event"
	"github.com/hekmekk/git-team/src/core/state_repository"
)

// Command the status command
func Command(root commandadapter.CommandRoot) *kingpin.CmdClause {
	status := root.Command("status", "Print the current status")

	status.Action(commandadapter.Run(Policy(), statuseventadapter.MapEventToEffects))

	return status
}

// Policy the status policy constructor
func Policy() status.Policy {
	return status.Policy{
		Deps: status.Dependencies{
			StateRepositoryQuery: staterepository.Query,
		},
	}
}
