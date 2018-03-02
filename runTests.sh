#!/bin/bash

PSQL_SERVICE=PostgreSQL
COMPOSE_FILE=docker-compose.yml
DATABASE_NAME=test
CONFIG_FILE=config.test.yml

# Install our project and output any errors
echo -n 'Building project...'
go install -v
echo 'finished'

# Create the tables in our database
echo 'Creating tables in DB...'
docker-compose -f ${COMPOSE_FILE} exec ${PSQL_SERVICE} sudo -u postgres psql -d ${DATABASE_NAME} -U kbase -f /docker-entrypoint-initdb.d/test/test.sql > /dev/null

# Run our server
echo 'Runnig knowlege base server'
knowledge-base -config=${CONFIG_FILE} > test_logs.txt 2>&1 &

PROJ_PID=$!

# Run our cucumber tests
echo 'Running cucumber tests...'
cucumber

# Run our unit tests
echo 'Running go unit tests...'
go test ./...

kill $PROJ_PID
