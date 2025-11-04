#!/bin/bash
#
# Build and Test Golang driver
#
# Usage: execute_tests.sh [TEST_NAME]
#   TEST_NAME: Optional test name pattern to pass to -run flag
#
set -e
set -o pipefail
CI_SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TOPDIR=$(cd $CI_SCRIPTS_DIR/../.. && pwd)
eval $(jq -r '.testconnection | to_entries | map("export \(.key)=\(.value|tostring)")|.[]' $TOPDIR/parameters.json)
env | grep SNOWFLAKE | grep -v PASS | grep -v SECRET | sort
cd $TOPDIR
go install github.com/jstemmer/go-junit-report/v2@latest

# Parse optional test name argument
TEST_RUN_FLAG=""
if [[ -n "$1" ]]; then
  TEST_RUN_FLAG="-run $1"
fi

if [[ "$HOME_EMPTY" == "yes" ]] ; then
  export GOCACHE=$HOME/go-build
  export GOMODCACHE=$HOME/go-modules
  export HOME=
fi
if [[ -n "$JENKINS_HOME" ]]; then
  export WORKSPACE=${WORKSPACE:-/mnt/workspace}
  go test $GO_TEST_PARAMS $TEST_RUN_FLAG -timeout 90m -race -v . | /home/user/go/bin/go-junit-report -iocopy -out $WORKSPACE/junit-go.xml
else
  set +e
  go test $GO_TEST_PARAMS $TEST_RUN_FLAG -timeout 90m -race -coverprofile=coverage.txt -covermode=atomic -v . | tee test-output.txt
  TEST_EXIT_CODE=$?
  cat test-output.txt | go-junit-report > test-report.junit.xml
  exit $TEST_EXIT_CODE
fi