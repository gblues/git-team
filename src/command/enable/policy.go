package enable

import (
	"os"
	"path"

	utils "github.com/hekmekk/git-team/src/command/enable/utils"
	"github.com/hekmekk/git-team/src/core/config"
	"github.com/hekmekk/git-team/src/core/events"
)

// Dependencies the dependencies of the enable Policy module
type Dependencies struct {
	SanityCheckCoauthors          func([]string) []error
	LoadConfig                    func() config.Config
	CreateTemplateDir             func(path string, perm os.FileMode) error
	WriteTemplateFile             func(path string, data []byte, mode os.FileMode) error
	GitSetCommitTemplate          func(path string) error
	GitSetHooksPath               func(path string) error
	GitResolveAliases             func(aliases []string) ([]string, []error)
	StateRepositoryPersistEnabled func(coauthors []string) error
}

// Request the coauthors with which to enable git-team
type Request struct {
	AliasesAndCoauthors *[]string
}

// Policy add a <Coauthor> under "team.alias.<Alias>"
type Policy struct {
	Deps Dependencies
	Req  Request
}

// Apply enable git-team with the provided co-authors
func (policy Policy) Apply() events.Event {
	deps := policy.Deps
	req := policy.Req

	aliasesAndCoauthors := append(*req.AliasesAndCoauthors) // should be == *req.AliasesAndCoauthors

	if len(aliasesAndCoauthors) == 0 {
		return Aborted{}
	}

	coauthors, errs := applyAdditionalGuards(deps, aliasesAndCoauthors)
	if len(errs) > 0 {
		return Failed{Reason: errs}
	}

	uniqueCoauthors := removeDuplicates(coauthors)

	cfg := deps.LoadConfig()

	if err := setupTemplate(deps, cfg, uniqueCoauthors); err != nil {
		return Failed{Reason: []error{err}}
	}

	if err := deps.GitSetHooksPath(cfg.GitTeamHooksPath); err != nil {
		return Failed{Reason: []error{err}}
	}

	if err := deps.StateRepositoryPersistEnabled(uniqueCoauthors); err != nil {
		return Failed{Reason: []error{err}}
	}

	return Succeeded{}
}

func applyAdditionalGuards(deps Dependencies, aliasesAndCoauthors []string) ([]string, []error) {
	coauthorCandidates, aliases := utils.Partition(aliasesAndCoauthors)

	sanityCheckErrs := deps.SanityCheckCoauthors(coauthorCandidates)
	if len(sanityCheckErrs) > 0 {
		return []string{}, sanityCheckErrs
	}

	resolvedAliases, resolveErrs := deps.GitResolveAliases(aliases)
	if len(resolveErrs) > 0 {
		return []string{}, resolveErrs
	}

	return append(coauthorCandidates, resolvedAliases...), []error{}
}

func removeDuplicates(coauthors []string) []string {
	var uniqueCoauthors []string
	temp := make(map[string]bool)
	for _, coauthor := range coauthors {
		temp[coauthor] = true
	}

	for coauthor := range temp {
		uniqueCoauthors = append(uniqueCoauthors, coauthor)
	}

	return uniqueCoauthors
}

func setupTemplate(deps Dependencies, cfg config.Config, uniqueCoauthors []string) error {
	commitTemplatePath := cfg.GitTeamCommitTemplatePath
	commitTemplateDir := path.Dir(commitTemplatePath)

	if err := deps.CreateTemplateDir(commitTemplateDir, os.ModePerm); err != nil {
		return err
	}

	if err := deps.WriteTemplateFile(commitTemplatePath, []byte(utils.PrepareForCommitMessage(uniqueCoauthors)), 0644); err != nil {
		return err
	}
	if err := deps.GitSetCommitTemplate(commitTemplatePath); err != nil {
		return err
	}

	return nil
}
