#!/bin/sh
context=$1
if [ -z "$context" ]; then
    context="eng-eks"
fi

SECRET="$(kubectl --context=$context -n eng get secret snowflake-sfdrivertest-credentials -o=json | jq -r '.data')"

# Create parameters.json
(
	echo -n '{"testconnection":{'
	for E in SNOWFLAKE_TEST_ACCOUNT SNOWFLAKE_TEST_DATABASE SNOWFLAKE_TEST_WAREHOUSE SNOWFLAKE_TEST_USER SNOWFLAKE_TEST_ROLE; do
		echo -n "$COMMA\"$E\":\""`echo ${SECRET} | jq -r --arg E "$E" '.["'$E'"]' | base64 --decode`"\""
		COMMA=,
	done
	echo -n ', "SNOWFLAKE_TEST_PRIVATE_KEY": "rsa-2048-private-key.pem"'
	echo -n ', "SNOWFLAKE_TEST_AUTHENTICATOR": "SNOWFLAKE_JWT"'
	echo '}}'
) | jq -r '.' > parameters.json

# Read JWT key
echo ${SECRET} | jq -r '.["SNOWFLAKE_TEST_JWT_PRIVATE_KEY"]' | base64 --decode > rsa-2048-private-key.p8

# Read TEST key
echo ${SECRET} | jq -r '.["SNOWFLAKE_TEST_PRIVATE_KEY"]' | base64 --decode > rsa-2048-private-key.pem
