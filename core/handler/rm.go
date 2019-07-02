package handler

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/hekmekk/git-team/core/git"
)

func RemoveCommand(alias *string) {
	_, resolveErr := git.ResolveAlias(*alias)
	if resolveErr != nil {
		color.Yellow(fmt.Sprintf("No such alias: '%s'.", *alias))
		os.Exit(0)
	}

	err := git.RemoveAlias(*alias)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error: %s\n", err.Error()))
		os.Exit(-1)
	}
	color.Red(fmt.Sprintf("Alias '%s' has been removed.", *alias))
}
