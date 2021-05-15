#!/usr/bin/env bats

load '/bats-libs/bats-support/load.bash'
load '/bats-libs/bats-assert/load.bash'

REPO_PATH=/tmp/repo/use-cases-repo-local
USER_NAME=git-team-acceptance-test
USER_EMAIL=acc@git.team

setup() {
	git config --global init.defaultBranch main

	mkdir -p $REPO_PATH
	cd $REPO_PATH
	touch THE_FILE

	git init
	git config user.name "$USER_NAME"
	git config user.email "$USER_EMAIL"

	/usr/local/bin/git-team config activation-scope repo-local
}

teardown() {
	/usr/local/bin/git-team disable

	cd -
	rm -rf $REPO_PATH

	rm /root/.gitconfig
}

@test "use case: (scope: repo-local) an existing repo-local git hook should be respected - commit msg" {
	echo -e '#!/bin/sh\necho "commit-msg hook triggered with params: $@"\nexit 1' > $REPO_PATH/.git/hooks/commit-msg
	chmod +x $REPO_PATH/.git/hooks/commit-msg

	/usr/local/bin/git-team enable 'A <a@x.y>'

	git add -A
	run git commit -m "test"

	assert_failure
	assert_line --index 0 'commit-msg hook triggered with params: .git/COMMIT_EDITMSG'
}

@test "use case: (scope: repo-local) an existing repo-local git hook should be respected - prepare-commit-msg" {
	echo -e '#!/bin/sh\necho "prepare-commit-msg hook triggered with params: $@"\nexit 1' > $REPO_PATH/.git/hooks/prepare-commit-msg
	chmod +x $REPO_PATH/.git/hooks/prepare-commit-msg

	/usr/local/bin/git-team enable 'A <a@x.y>'

	git add -A
	run git commit -m "test"

	assert_failure
	assert_line --index 0 'prepare-commit-msg hook triggered with params: .git/COMMIT_EDITMSG message'
}

@test "use case: (scope: repo-local) when git-team is enabled then 'git commit -m' should have the respective co-authors injected" {
	/usr/local/bin/git-team enable 'B <b@x.y>' 'A <a@x.y>' 'C <c@x.y>'

	git add -A
	git commit -m "test"

	run git show --name-only
	assert_success
	assert_line --index 0 --regexp '^commit\s\w+'
	assert_line --index 1 "Author: $USER_NAME <$USER_EMAIL>"
	assert_line --index 2 --regexp '^Date:.+'
	assert_line --index 3 --regexp '\s+test'
	refute_line --index 4 --regexp '\w+'
	assert_line --index 5 --regexp '\s+Co-authored-by: A <a@x.y>'
	assert_line --index 6 --regexp '\s+Co-authored-by: B <b@x.y>'
	assert_line --index 7 --regexp '\s+Co-authored-by: C <c@x.y>'
	assert_line --index 8 'THE_FILE'
}

@test "use case: (scope: repo-local) when git-team is enabled then 'git commit -m' should not result in interference with existing co-authors" {
	/usr/local/bin/git-team enable 'B <b@x.y>' 'A <a@x.y>' 'C <c@x.y>'

	git add -A
	git commit -F- <<EOF
test

Co-authored-by: D <d@x.y>
EOF

	run git show --name-only
	assert_success
	assert_line --index 0 --regexp '^commit\s\w+'
	assert_line --index 1 "Author: $USER_NAME <$USER_EMAIL>"
	assert_line --index 2 --regexp '^Date:.+'
	assert_line --index 3 --regexp '\s+test'
	refute_line --index 4 --regexp '\w+'
	assert_line --index 5 --regexp '\s+Co-authored-by: D <d@x.y>'
	assert_line --index 6 'THE_FILE'
}

@test "use case: (scope: repo-local) when git-team is disabled then 'git commit -m' should not have any co-authors injected" {
	/usr/local/bin/git-team disable

	git add -A
	git commit -m "test"

	run git show --name-only
	assert_success
	assert_line --index 0 --regexp '^commit\s\w+'
	assert_line --index 1 "Author: $USER_NAME <$USER_EMAIL>"
	assert_line --index 2 --regexp '^Date:.+'
	assert_line --index 3 --regexp '\s+test'
	assert_line --index 4 'THE_FILE'
	refute_output --partial 'Co-authored-by:'
}
