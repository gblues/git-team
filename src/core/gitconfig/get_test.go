package gitconfig

import (
	"errors"
	"testing"
)

func TestShouldReturnTheValue(t *testing.T) {
	expectedValue := "value"
	execSucceeds := func(args ...string) ([]string, error) { return []string{expectedValue}, nil }

	value, err := get(execSucceeds)("key")

	if err != nil {
		t.Errorf("expected err: %s, got err: %s", "[no err]", err)
		t.Fail()
	}

	if expectedValue != value {
		t.Errorf("expected: %s, received: %s", expectedValue, value)
		t.Fail()
	}
}

func TestShouldReturnEmptyStringIfNoValueIsFound(t *testing.T) {
	execSucceedsEmpty := func(args ...string) ([]string, error) { return []string{}, nil }

	value, err := get(execSucceedsEmpty)("key")

	if err != nil {
		t.Errorf("expected err: %s, got err: %s", "[no err]", err)
		t.Fail()
	}

	if "" != value {
		t.Errorf("expected: %s, received: %s", "", value)
		t.Fail()
	}
}

func TestShouldReturnTheErrorIfAnErrorOccurs(t *testing.T) {
	expectedErr := errors.New("git command failed")
	execFails := func(args ...string) ([]string, error) { return nil, expectedErr }

	value, err := get(execFails)("key")

	if expectedErr != err {
		t.Errorf("expected err: %s, got err: %s", expectedErr, err)
		t.Fail()
	}

	if "" != value {
		t.Errorf("expected: %s, received: %s", "", value)
		t.Fail()
	}
}
