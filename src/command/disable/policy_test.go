package disable

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	activationscope "github.com/hekmekk/git-team/src/shared/activation/scope"
	config "github.com/hekmekk/git-team/src/shared/config/entity/config"
	gitconfigerror "github.com/hekmekk/git-team/src/shared/gitconfig/error"
	gitconfigscope "github.com/hekmekk/git-team/src/shared/gitconfig/scope"
)

type gitConfigReaderMock struct {
	get func(gitconfigscope.Scope, string) (string, error)
}

func (mock gitConfigReaderMock) Get(scope gitconfigscope.Scope, key string) (string, error) {
	return mock.get(scope, key)
}

func (mock gitConfigReaderMock) GetAll(scope gitconfigscope.Scope, key string) ([]string, error) {
	return []string{}, nil
}

func (mock gitConfigReaderMock) GetRegexp(scope gitconfigscope.Scope, pattern string) (map[string]string, error) {
	return nil, nil
}

func (mock gitConfigReaderMock) List(scope gitconfigscope.Scope) (map[string]string, error) {
	return nil, nil
}

type gitConfigWriterMock struct {
	unsetAll func(gitconfigscope.Scope, string) error
}

func (mock gitConfigWriterMock) UnsetAll(scope gitconfigscope.Scope, key string) error {
	return mock.unsetAll(scope, key)
}

func (mock gitConfigWriterMock) ReplaceAll(scope gitconfigscope.Scope, key string, value string) error {
	return nil
}

func (mock gitConfigWriterMock) Add(scope gitconfigscope.Scope, key string, value string) error {
	return nil
}

type stateWriterMock struct {
	persistDisabled func(activationscope.Scope) error
}

func (mock stateWriterMock) PersistEnabled(scope activationscope.Scope, coauthors []string, previousHooksPath string) error {
	return nil
}
func (mock stateWriterMock) PersistDisabled(scope activationscope.Scope) error {
	return mock.persistDisabled(scope)
}

type configReaderMock struct {
	read func() (config.Config, error)
}

func (mock configReaderMock) Read() (config.Config, error) {
	return mock.read()
}

type activationValidatorMock struct {
	isInsideAGitRepository func() bool
}

func (mock activationValidatorMock) IsInsideAGitRepository() bool {
	return mock.isInsideAGitRepository()
}

var (
	fileInfo        os.FileInfo
	statFile        = func(string) (os.FileInfo, error) { return fileInfo, nil }
	removeFile      = func(string) error { return nil }
	persistDisabled = func() error { return nil }
)

func TestDisableSucceeds(t *testing.T) {
	t.Parallel()

	templatePath := "/path/to/template"

	deps := Dependencies{
		StatFile: statFile,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Succeeded{}

	cases := []struct {
		activationScope      activationscope.Scope
		gitconfigScope       gitconfigscope.Scope
		deletionTemplatePath string
	}{
		{activationscope.Global, gitconfigscope.Global, "/path/to/template"},
		{activationscope.RepoLocal, gitconfigscope.Local, "/path/to"},
	}

	for _, caseLoopVar := range cases {
		activationScope := caseLoopVar.activationScope
		expectedGitConfigScope := caseLoopVar.gitconfigScope
		expectedPathToDelete := caseLoopVar.deletionTemplatePath

		t.Run(activationScope.String(), func(t *testing.T) {
			t.Parallel()

			deps.ConfigReader = &configReaderMock{
				read: func() (config.Config, error) {
					return config.Config{ActivationScope: activationScope}, nil
				},
			}

			deps.GitConfigReader = &gitConfigReaderMock{
				get: func(scope gitconfigscope.Scope, key string) (string, error) {
					if scope != expectedGitConfigScope {
						t.Errorf("wrong scope, expected: %s, got: %s", expectedGitConfigScope, scope)
						t.Fail()
					}
					return templatePath, nil
				},
			}

			deps.GitConfigWriter = &gitConfigWriterMock{
				unsetAll: func(scope gitconfigscope.Scope, key string) error {
					if key != "commit.template" && key != "core.hooksPath" {
						t.Errorf("wrong key: %s", key)
						t.Fail()
					}
					if scope != expectedGitConfigScope {
						t.Errorf("wrong scope, expected: %s, got: %s", expectedGitConfigScope, scope)
						t.Fail()
					}
					return nil
				},
			}

			deps.StateWriter = &stateWriterMock{
				persistDisabled: func(scope activationscope.Scope) error {
					if scope != activationScope {
						t.Errorf("wrong scope, expected: %s, got: %s", activationScope, scope)
						t.Fail()
					}
					return nil
				},
			}

			deps.RemoveFile = func(path string) error {
				if path != expectedPathToDelete {
					t.Errorf("trying to delete the wrong commit template file, expected: %s, got: %s", expectedPathToDelete, path)
					t.Fail()
				}
				return nil
			}

			event := Policy{deps}.Apply()

			if !reflect.DeepEqual(expectedEvent, event) {
				t.Errorf("expected: %s, got: %s", expectedEvent, event)
				t.Fail()
			}
		})
	}
}

func TestDisableShouldSucceedWhenUnsetHooksPathFailsBecauseTheOptionDoesntExist(t *testing.T) {
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "/path/to/template", nil
		},
	}

	expectedErr := gitconfigerror.ErrTryingToUnsetAnOptionWhichDoesNotExist
	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return expectedErr
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	stateWriter := &stateWriterMock{
		persistDisabled: func(scope activationscope.Scope) error {
			return nil
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		StatFile:        statFile,
		RemoveFile:      removeFile,
		StateWriter:     stateWriter,
		ConfigReader:    configReader,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Succeeded{}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldFailWhenNotInsideAGitRepository(t *testing.T) {
	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.RepoLocal}, nil
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigWriter: gitConfigWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return false
			},
		},
	}

	expectedErr := errors.New("failed to disable with activation-scope=repo-local: not inside a git repository")

	expectedEvent := Failed{Reason: expectedErr}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldFailWhenReadConfigFails(t *testing.T) {
	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{}, gitconfigerror.ErrConfigFileIsInvalid
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigWriter: gitConfigWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Failed{Reason: fmt.Errorf("failed to read config: %s", gitconfigerror.ErrConfigFileIsInvalid)}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldFailWhenUnsetHooksPathFails(t *testing.T) {
	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return gitconfigerror.ErrConfigFileCannotBeWritten
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigWriter: gitConfigWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Failed{Reason: fmt.Errorf("failed to unset core.hooksPath: %s", gitconfigerror.ErrConfigFileCannotBeWritten)}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldSucceedWhenReadingCommitTemplateFailsBecauseItHasBeenUnsetAlready(t *testing.T) {
	expectedErr := gitconfigerror.ErrSectionOrKeyIsInvalid
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "", expectedErr
		},
	}

	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	stateWriter := &stateWriterMock{
		persistDisabled: func(scope activationscope.Scope) error {
			return nil
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		StatFile:        statFile,
		RemoveFile:      removeFile,
		StateWriter:     stateWriter,
		ConfigReader:    configReader,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Succeeded{}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldFailWhenReadingCommitTemplatePathFails(t *testing.T) {
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "", gitconfigerror.ErrConfigFileCannotBeWritten
		},
	}

	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Failed{Reason: fmt.Errorf("failed to get commit.template: %s", gitconfigerror.ErrConfigFileCannotBeWritten)}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldSucceedWhenUnsetCommitTemplateFailsBecauseItWasUnsetAlready(t *testing.T) {
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "/path/to/template", nil
		},
	}

	expectedErr := gitconfigerror.ErrTryingToUnsetAnOptionWhichDoesNotExist
	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return expectedErr
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	stateWriter := &stateWriterMock{
		persistDisabled: func(scope activationscope.Scope) error {
			return nil
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		StatFile:        statFile,
		RemoveFile:      removeFile,
		StateWriter:     stateWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Succeeded{}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldFailWhenUnsetCommitTemplateFails(t *testing.T) {
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "/path/to/template", nil
		},
	}

	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return gitconfigerror.ErrConfigFileCannotBeWritten
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Failed{Reason: fmt.Errorf("failed to unset commit.template: %s", gitconfigerror.ErrConfigFileCannotBeWritten)}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldFailWhenRemoveFileFails(t *testing.T) {
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "/path/to/template", nil
		},
	}

	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	removeFileErr := errors.New("failed to remove file")

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		StatFile:        statFile,
		RemoveFile:      func(string) error { return removeFileErr },
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Failed{Reason: fmt.Errorf("failed to remove commit template: %s", removeFileErr)}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldSucceedButNotTryToRemoveTheCommitTemplateFileWhenTheRespectivePathIsEmpty(t *testing.T) {
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "", nil
		},
	}

	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	stateWriter := &stateWriterMock{
		persistDisabled: func(scope activationscope.Scope) error {
			return nil
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		StatFile:        func(string) (os.FileInfo, error) { return fileInfo, nil },
		StateWriter:     stateWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Succeeded{}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldSucceedButNotTryToRemoveTheCommitTemplateFileWhenStatFileFails(t *testing.T) {
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "/path/to/template", nil
		},
	}

	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	stateWriter := &stateWriterMock{
		persistDisabled: func(scope activationscope.Scope) error {
			return nil
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		StatFile:        func(string) (os.FileInfo, error) { return fileInfo, errors.New("failed to stat file") },
		StateWriter:     stateWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Succeeded{}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}

func TestDisableShouldFailWhenpersistDisabledFails(t *testing.T) {
	gitConfigReader := &gitConfigReaderMock{
		get: func(_ gitconfigscope.Scope, key string) (string, error) {
			return "/path/to/template", nil
		},
	}

	gitConfigWriter := &gitConfigWriterMock{
		unsetAll: func(scope gitconfigscope.Scope, key string) error {
			switch key {
			case "commit.template":
				return nil
			case "core.hooksPath":
				return nil
			default:
				return fmt.Errorf("wrong key: %s", key)
			}
		},
	}

	writeStateErr := errors.New("failed to save status")

	stateWriter := &stateWriterMock{
		persistDisabled: func(scope activationscope.Scope) error {
			return writeStateErr
		},
	}

	configReader := &configReaderMock{
		read: func() (config.Config, error) {
			return config.Config{ActivationScope: activationscope.Global}, nil
		},
	}

	deps := Dependencies{
		ConfigReader:    configReader,
		GitConfigReader: gitConfigReader,
		GitConfigWriter: gitConfigWriter,
		StatFile:        statFile,
		RemoveFile:      removeFile,
		StateWriter:     stateWriter,
		ActivationValidator: &activationValidatorMock{
			isInsideAGitRepository: func() bool {
				return true
			},
		},
	}

	expectedEvent := Failed{Reason: fmt.Errorf("failed to write current state: %s", writeStateErr)}

	event := Policy{deps}.Apply()

	if !reflect.DeepEqual(expectedEvent, event) {
		t.Errorf("expected: %s, got: %s", expectedEvent, event)
		t.Fail()
	}
}
