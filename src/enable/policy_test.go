package enable

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/hekmekk/git-team/src/core/config"
)

func TestEnableAborted(t *testing.T) {
	deps := Dependencies{}
	req := Request{AliasesAndCoauthors: &[]string{}}

	expectedEvent := Aborted{}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableSucceeds(t *testing.T) {
	coauthors := &[]string{"Mr. Noujz <noujz@mr.se>", "mrs"}
	expectedStateRepositoryPersistEnabledCoauthors := []string{"Mr. Noujz <noujz@mr.se>", "Mrs. Noujz <noujz@mrs.se>"}
	expectedCommitTemplateCoauthors := "\n\nCo-authored-by: Mr. Noujz <noujz@mr.se>\nCo-authored-by: Mrs. Noujz <noujz@mrs.se>"

	setHooksPath := func(string) error { return nil }
	CreateTemplateDir := func(string, os.FileMode) error { return nil }
	WriteTemplateFile := func(_ string, data []byte, _ os.FileMode) error {
		if expectedCommitTemplateCoauthors != string(data) {
			t.Errorf("expected: %s, got: %s", expectedCommitTemplateCoauthors, string(data))
			t.Fail()
		}
		return nil
	}
	GitSetCommitTemplate := func(string) error { return nil }
	resolveAliases := func([]string) ([]string, []error) { return []string{"Mrs. Noujz <noujz@mrs.se>"}, []error{} }
	StateRepositoryPersistEnabled := func(coauthors []string) error {
		if !reflect.DeepEqual(expectedStateRepositoryPersistEnabledCoauthors, coauthors) {
			t.Errorf("expected: %s, got: %s", expectedStateRepositoryPersistEnabledCoauthors, coauthors)
			t.Fail()
		}
		return nil
	}
	cfg := config.Config{TemplateFileName: "TEMPLATE_FILE", BaseDir: "BASE_DIR", StatusFileName: "STATUS_FILE"}
	loadConfig := func() (config.Config, error) { return cfg, nil }

	deps := Dependencies{
		SanityCheckCoauthors:          func(coauthors []string) []error { return []error{} },
		CreateTemplateDir:             CreateTemplateDir,
		WriteTemplateFile:             WriteTemplateFile,
		GitSetHooksPath:               setHooksPath,
		GitSetCommitTemplate:          GitSetCommitTemplate,
		GitResolveAliases:             resolveAliases,
		StateRepositoryPersistEnabled: StateRepositoryPersistEnabled,
		LoadConfig:                    loadConfig,
	}
	req := Request{AliasesAndCoauthors: coauthors}

	expectedEvent := Succeeded{}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableDropsDuplicateEntries(t *testing.T) {
	coauthors := []string{"Mr. Noujz <noujz@mr.se>", "mrs", "Mr. Noujz <noujz@mr.se>", "mrs", "Mrs. Noujz <noujz@mrs.se>"}
	expectedStateRepositoryPersistEnabledCoauthors := []string{"Mr. Noujz <noujz@mr.se>", "Mrs. Noujz <noujz@mrs.se>"}
	expectedCommitTemplateCoauthors := "\n\nCo-authored-by: Mr. Noujz <noujz@mr.se>\nCo-authored-by: Mrs. Noujz <noujz@mrs.se>"

	setHooksPath := func(string) error { return nil }
	CreateTemplateDir := func(string, os.FileMode) error { return nil }
	WriteTemplateFile := func(_ string, data []byte, _ os.FileMode) error {
		if expectedCommitTemplateCoauthors != string(data) {
			t.Errorf("expected: %s, got: %s", expectedCommitTemplateCoauthors, string(data))
			t.Fail()
		}
		return nil
	}
	GitSetCommitTemplate := func(string) error { return nil }
	resolveAliases := func([]string) ([]string, []error) { return []string{"Mrs. Noujz <noujz@mrs.se>"}, []error{} }
	StateRepositoryPersistEnabled := func(coauthors []string) error {
		if !reflect.DeepEqual(expectedStateRepositoryPersistEnabledCoauthors, coauthors) {
			t.Errorf("expected: %s, got: %s", expectedStateRepositoryPersistEnabledCoauthors, coauthors)
			t.Fail()
		}
		return nil
	}
	cfg := config.Config{TemplateFileName: "TEMPLATE_FILE", BaseDir: "BASE_DIR", StatusFileName: "STATUS_FILE"}
	loadConfig := func() (config.Config, error) { return cfg, nil }

	deps := Dependencies{
		SanityCheckCoauthors:          func(coauthors []string) []error { return []error{} },
		CreateTemplateDir:             CreateTemplateDir,
		WriteTemplateFile:             WriteTemplateFile,
		GitSetHooksPath:               setHooksPath,
		GitSetCommitTemplate:          GitSetCommitTemplate,
		GitResolveAliases:             resolveAliases,
		StateRepositoryPersistEnabled: StateRepositoryPersistEnabled,
		LoadConfig:                    loadConfig,
	}
	req := Request{AliasesAndCoauthors: &coauthors}

	expectedEvent := Succeeded{}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableFailsDueToSanityCheckErr(t *testing.T) {
	coauthors := []string{"INVALID COAUTHOR"}

	expectedErr := errors.New("Not a valid coauthor: INVALID COAUTHOR")

	deps := Dependencies{
		SanityCheckCoauthors: func(coauthors []string) []error { return []error{expectedErr} },
	}
	req := Request{AliasesAndCoauthors: &coauthors}

	expectedEvent := Failed{Reason: []error{expectedErr}}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableFailsDueToResolveAliasesErr(t *testing.T) {
	coauthors := []string{"Mr. Noujz <noujz@mr.se>", "mrs"}

	expectedErr := errors.New("failed to resolve alias mrs")

	deps := Dependencies{
		SanityCheckCoauthors: func(coauthors []string) []error { return []error{} },
		GitResolveAliases:    func([]string) ([]string, []error) { return []string{}, []error{expectedErr} },
	}
	req := Request{AliasesAndCoauthors: &coauthors}

	expectedEvent := Failed{Reason: []error{expectedErr}}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableFailsDueToLoadConfigErr(t *testing.T) {
	coauthors := []string{"Mr. Noujz <noujz@mr.se>"}

	expectedErr := errors.New("failed to load config")

	sanityCheck := func([]string) []error { return []error{} }
	resolveAliases := func([]string) ([]string, []error) { return []string{"Mrs. Noujz <noujz@mrs.se>"}, []error{} }
	loadConfig := func() (config.Config, error) { return config.Config{}, expectedErr }

	deps := Dependencies{
		SanityCheckCoauthors: sanityCheck,
		GitResolveAliases:    resolveAliases,
		LoadConfig:           loadConfig,
	}
	req := Request{AliasesAndCoauthors: &coauthors}

	expectedEvent := Failed{Reason: []error{expectedErr}}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableFailsDueToCreateTemplateDirErr(t *testing.T) {
	coauthors := []string{"Mr. Noujz <noujz@mr.se>"}

	expectedErr := errors.New("Failed to create Dir")

	sanityCheck := func([]string) []error { return []error{} }
	resolveAliases := func([]string) ([]string, []error) { return []string{"Mrs. Noujz <noujz@mrs.se>"}, []error{} }
	cfg := config.Config{TemplateFileName: "TEMPLATE_FILE", BaseDir: "BASE_DIR", StatusFileName: "STATUS_FILE"}
	loadConfig := func() (config.Config, error) { return cfg, nil }
	CreateTemplateDir := func(string, os.FileMode) error { return expectedErr }

	deps := Dependencies{
		SanityCheckCoauthors: sanityCheck,
		CreateTemplateDir:    CreateTemplateDir,
		GitResolveAliases:    resolveAliases,
		LoadConfig:           loadConfig,
	}
	req := Request{AliasesAndCoauthors: &coauthors}

	expectedEvent := Failed{Reason: []error{expectedErr}}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableFailsDueToWriteTemplateFileErr(t *testing.T) {
	coauthors := &[]string{"Mr. Noujz <noujz@mr.se>"}

	expectedErr := errors.New("Failed to write file")

	sanityCheck := func([]string) []error { return []error{} }
	resolveAliases := func([]string) ([]string, []error) { return []string{"Mrs. Noujz <noujz@mrs.se>"}, []error{} }
	cfg := config.Config{TemplateFileName: "TEMPLATE_FILE", BaseDir: "BASE_DIR", StatusFileName: "STATUS_FILE"}
	loadConfig := func() (config.Config, error) { return cfg, nil }
	CreateTemplateDir := func(string, os.FileMode) error { return nil }
	WriteTemplateFile := func(string, []byte, os.FileMode) error { return expectedErr }

	deps := Dependencies{
		SanityCheckCoauthors: sanityCheck,
		CreateTemplateDir:    CreateTemplateDir,
		WriteTemplateFile:    WriteTemplateFile,
		GitResolveAliases:    resolveAliases,
		LoadConfig:           loadConfig,
	}
	req := Request{AliasesAndCoauthors: coauthors}

	expectedEvent := Failed{Reason: []error{expectedErr}}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableFailsDueToGitSetCommitTemplateErr(t *testing.T) {
	coauthors := &[]string{"Mr. Noujz <noujz@mr.se>"}

	expectedErr := errors.New("Failed to set commit template")

	sanityCheck := func([]string) []error { return []error{} }
	resolveAliases := func([]string) ([]string, []error) { return []string{"Mrs. Noujz <noujz@mrs.se>"}, []error{} }
	cfg := config.Config{TemplateFileName: "TEMPLATE_FILE", BaseDir: "BASE_DIR", StatusFileName: "STATUS_FILE"}
	loadConfig := func() (config.Config, error) { return cfg, nil }
	CreateTemplateDir := func(string, os.FileMode) error { return nil }
	WriteTemplateFile := func(string, []byte, os.FileMode) error { return nil }
	GitSetCommitTemplate := func(string) error { return expectedErr }

	deps := Dependencies{
		SanityCheckCoauthors: sanityCheck,
		CreateTemplateDir:    CreateTemplateDir,
		WriteTemplateFile:    WriteTemplateFile,
		GitSetCommitTemplate: GitSetCommitTemplate,
		GitResolveAliases:    resolveAliases,
		LoadConfig:           loadConfig,
	}
	req := Request{AliasesAndCoauthors: coauthors}

	expectedEvent := Failed{Reason: []error{expectedErr}}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableFailsDueToSetHooksPathErr(t *testing.T) {
	coauthors := &[]string{"Mr. Noujz <noujz@mr.se>"}

	expectedErr := errors.New("Failed to set hooks path")

	sanityCheck := func([]string) []error { return []error{} }
	resolveAliases := func([]string) ([]string, []error) { return []string{"Mrs. Noujz <noujz@mrs.se>"}, []error{} }
	cfg := config.Config{TemplateFileName: "TEMPLATE_FILE", BaseDir: "BASE_DIR", StatusFileName: "STATUS_FILE"}
	loadConfig := func() (config.Config, error) { return cfg, nil }
	CreateTemplateDir := func(string, os.FileMode) error { return nil }
	WriteTemplateFile := func(string, []byte, os.FileMode) error { return nil }
	GitSetCommitTemplate := func(string) error { return nil }
	setHooksPath := func(string) error { return expectedErr }

	deps := Dependencies{
		SanityCheckCoauthors: sanityCheck,
		CreateTemplateDir:    CreateTemplateDir,
		WriteTemplateFile:    WriteTemplateFile,
		GitSetHooksPath:      setHooksPath,
		GitSetCommitTemplate: GitSetCommitTemplate,
		GitResolveAliases:    resolveAliases,
		LoadConfig:           loadConfig,
	}
	req := Request{AliasesAndCoauthors: coauthors}

	expectedEvent := Failed{Reason: []error{expectedErr}}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestEnableFailsDueToSaveStatusErr(t *testing.T) {
	coauthors := &[]string{"Mr. Noujz <noujz@mr.se>"}

	expectedErr := errors.New("Failed to set status")

	sanityCheck := func([]string) []error { return []error{} }
	resolveAliases := func([]string) ([]string, []error) { return []string{"Mrs. Noujz <noujz@mrs.se>"}, []error{} }
	cfg := config.Config{TemplateFileName: "TEMPLATE_FILE", BaseDir: "BASE_DIR", StatusFileName: "STATUS_FILE"}
	loadConfig := func() (config.Config, error) { return cfg, nil }
	CreateTemplateDir := func(string, os.FileMode) error { return nil }
	WriteTemplateFile := func(string, []byte, os.FileMode) error { return nil }
	GitSetCommitTemplate := func(string) error { return nil }
	setHooksPath := func(string) error { return expectedErr }

	deps := Dependencies{
		SanityCheckCoauthors:          sanityCheck,
		CreateTemplateDir:             CreateTemplateDir,
		WriteTemplateFile:             WriteTemplateFile,
		GitSetHooksPath:               setHooksPath,
		GitSetCommitTemplate:          GitSetCommitTemplate,
		GitResolveAliases:             resolveAliases,
		LoadConfig:                    loadConfig,
		StateRepositoryPersistEnabled: func([]string) error { return expectedErr },
	}

	req := Request{AliasesAndCoauthors: coauthors}

	expectedEvent := Failed{Reason: []error{expectedErr}}

	event := Policy{deps, req}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func assertEqualsErr(t *testing.T, expectedErr error, errs []error) {
	if len(errs) != 1 {
		t.Errorf("got unexpected errs: %s", errs)
		t.Fail()
		return
	}
	if errs[0].Error() != expectedErr.Error() {
		t.Errorf("expexted: %s, got: %s", expectedErr.Error(), errs[0].Error())
		t.Fail()
		return
	}
}
