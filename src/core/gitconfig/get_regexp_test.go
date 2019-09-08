package gitconfig

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestShouldReturnTheKeyValueMap(t *testing.T) {
	key1 := "key1"
	val1 := "space separated value 1"
	key2 := "key2"
	val2 := "space separated value 2"

	expectedMap := make(map[string]string)
	expectedMap[key1] = val1
	expectedMap[key2] = val2

	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("%s %s", key1, val1))
	lines = append(lines, fmt.Sprintf("%s %s", key2, val2))

	execSucceeds := func(args ...string) ([]string, error) { return lines, nil }

	aliasMap, _ := getRegexp(execSucceeds)("pattern")

	if !reflect.DeepEqual(aliasMap, expectedMap) {
		t.Errorf("expected: %s, got: %s", expectedMap, aliasMap)
		t.Fail()
	}
}

func TestShouldReturnEmptyMapIfEmptyReturnFromGitConfigCommand(t *testing.T) {
	expectedMap := make(map[string]string, 0)
	execSucceedsEmpty := func(args ...string) ([]string, error) { return nil, nil }

	aliasMap, _ := getRegexp(execSucceedsEmpty)("pattern")

	if !reflect.DeepEqual(expectedMap, aliasMap) {
		t.Errorf("expected: %s, got: %s", expectedMap, aliasMap)
		t.Fail()
	}
}

func TestShouldFailIfGitConfigCommandFails(t *testing.T) {
	expectedMap := make(map[string]string, 0)
	expectedErr := errors.New("failed to exec git config command")

	execFails := func(args ...string) ([]string, error) { return nil, expectedErr }

	aliasMap, err := getRegexp(execFails)("pattern")

	if !reflect.DeepEqual(expectedMap, aliasMap) {
		t.Errorf("expected: %s, got: %s", expectedMap, aliasMap)
		t.Fail()
	}

	if expectedErr != err {
		t.Errorf("expected err: %s, got err: %s", expectedErr, err)
		t.Fail()
	}
}
