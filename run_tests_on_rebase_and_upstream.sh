#!/bin/bash

if [[ -z "$1" ]]; then
	echo "Error: No git commit hash of upstream provided"
	echo "Usage: $0 <commit-hash-of-upstream>"
	exit 1
fi

COMMIT_HASH="$1"

make test >|gosnowflake-test-rebase.txt 2>&1
git checkout "${COMMIT_HASH}"
go mod vendor
make test >|gosnowflake-test-upstream.txt 2>&1
