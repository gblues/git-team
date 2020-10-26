package enablecmdadapter

import (
	"io/ioutil"
	"os"

	"github.com/urfave/cli/v2"

	commandadapter "github.com/hekmekk/git-team/src/command/adapter"
	"github.com/hekmekk/git-team/src/command/enable"
	commitsettingsds "github.com/hekmekk/git-team/src/command/enable/commitsettings/datasource"
	enableeventadapter "github.com/hekmekk/git-team/src/command/enable/interfaceadapter/event"
	statuscmdmapper "github.com/hekmekk/git-team/src/command/status/interfaceadapter/cmd"
	"github.com/hekmekk/git-team/src/core/validation"
	activation "github.com/hekmekk/git-team/src/shared/activation/impl"
	configds "github.com/hekmekk/git-team/src/shared/config/datasource"
	gitconfig "github.com/hekmekk/git-team/src/shared/gitconfig/impl"
	gitconfiglegacy "github.com/hekmekk/git-team/src/shared/gitconfig/impl/legacy"
	state "github.com/hekmekk/git-team/src/shared/state/impl"
)

// Command the enable command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "enable",
		Usage: "Enables injection of the provided co-authors whenever `git-commit` is used",
		Before: func(c *cli.Context) error {
			// enable.PreAction(func(c *kingpin.ParseContext) error {
			// index := c.Peek().Index
			// numElements := len(c.Elements)
			// if index == 1 && numElements == 1 {
			// effects.NewDeprecationWarning("git team enable (without aliases)", "git team [status]").Run()
			// }
			// if index >= 1 && numElements == index+1 {
			// effects.NewDeprecationWarning("git team (without further sub-command specification)", "git team enable").Run()
			// }
			// return nil
			// })
			return nil
		},
		ArgsUsage: "co-authors - The co-authors for the next commit(s). A co-author must either be an alias or of the shape \"Name <email>\"",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all", Value: false, Aliases: []string{"A"}},
		},
		Action: func(c *cli.Context) error {
			coauthors := c.Args().Slice()
			useAll := c.Bool("all")
			return commandadapter.RunUrFave(policy(&coauthors, &useAll), enableeventadapter.MapEventToEffectsFactory(statuscmdmapper.Policy()))(c)
		},
	}
}

func policy(coauthors *[]string, useAll *bool) enable.Policy {
	return enable.Policy{
		Req: enable.Request{
			AliasesAndCoauthors: coauthors,
			UseAll:              useAll,
		},
		Deps: enable.Dependencies{
			SanityCheckCoauthors: validation.SanityCheckCoauthors,
			CreateTemplateDir:    os.MkdirAll,
			WriteTemplateFile:    ioutil.WriteFile,
			GitConfigWriter:      gitconfig.NewDataSink(),
			GitResolveAliases:    commandadapter.ResolveAliases,
			GitGetAssignments:    func() (map[string]string, error) { return gitconfiglegacy.GetRegexp("team.alias") },
			CommitSettingsReader: commitsettingsds.NewStaticValueDataSource(),
			ConfigReader:         configds.NewGitconfigDataSource(gitconfig.NewDataSource()),
			StateWriter:          state.NewGitConfigDataSink(gitconfig.NewDataSink()),
			GetEnv:               os.Getenv,
			GetWd:                os.Getwd,
			ActivationValidator:  activation.NewGitConfigDataSource(gitconfig.NewDataSource()),
		},
	}
}
