package datasource

import (
	"fmt"

	activationscope "github.com/hekmekk/git-team/src/command/config/entity/activationscope"
	config "github.com/hekmekk/git-team/src/command/config/entity/config"
	coregitconfig "github.com/hekmekk/git-team/src/core/gitconfig"
	giterror "github.com/hekmekk/git-team/src/core/gitconfig/error"
)

// GitconfigDataSource reads configuration from git config
type GitconfigDataSource struct {
	GitSettingsReader coregitconfig.SettingsReader
}

// NewGitconfigDataSource constructs new GitconfigDataSource
func NewGitconfigDataSource() GitconfigDataSource {
	return newGitconfigDataSource(coregitconfig.NewDataSource())
}

// for tests
func newGitconfigDataSource(gitSettingsReader coregitconfig.SettingsReader) GitconfigDataSource {
	return GitconfigDataSource{gitSettingsReader}
}

func (ds GitconfigDataSource) Read() (config.Config, error) {
	rawScope, err := ds.GitSettingsReader.Get("team.config.activation-scope")

	if err != nil && err.Error() == giterror.SectionOrKeyIsInvalid {
		return config.Config{ActivationScope: activationscope.Global}, nil
	}

	if err != nil {
		return config.Config{}, err
	}

	scope := activationscope.FromString(rawScope)
	if scope == activationscope.Unknown {
		return config.Config{}, fmt.Errorf("Unknown activation-scope '%s' found in config. Did you edit it manually?", rawScope)
	}

	cfg := config.Config{
		ActivationScope: scope,
	}

	return cfg, nil
}
