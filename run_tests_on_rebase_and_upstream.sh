#!/bin/bash

if [[ -z "$1" ]]; then
	echo "Error: No git commit hash of upstream provided"
	echo "Usage: $0 <commit-hash-of-upstream>"
	exit 1
fi

COMMIT_HASH="$1"

go mod vendor
make test >|gosnowflake-test-rebase.txt 2>&1
git checkout "${COMMIT_HASH}"
go mod vendor
make test >|gosnowflake-test-upstream.txt 2>&1

# write the diff of failed tests to a file
diff <(grep -r "FAIL:" gosnowflake-test-upstream.txt | sed -e 's/.*\(Test[^ ]*\) .*/\1/gm') <(grep -r "FAIL:" gosnowflake-test-rebase.txt | sed -e 's/.*\(Test[^ ]*\) .*/\1/gm') > diff_failed_tests.txt
