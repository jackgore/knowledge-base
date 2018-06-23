#!/bin/bash

PSQL_SERVICE=PostgreSQL
COMPOSE_FILE=docker-compose.yml
DATABASE_NAME=test
CONFIG_FILE=config.test.yml
KB_HOST=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | cut -d\  -f2)

# Install our project and output any errors
echo -n 'Building project...'
go install -v
echo 'finished'

# Create the tables in our database
echo 'Creating tables in DB...'
cat data/clearTables.sql data/init.sql | PGPASSWORD=password psql -U kbase -d test -h ${KB_HOST} -f - > /dev/null 2>&1

# Run our server
echo 'Running knowledge-base server'
knowledge-base -config=${CONFIG_FILE} > test_logs.txt 2>&1 &

PROJ_PID=$!

sleep 1

# Run our api-check tests
echo 'Running api-check tests...'
api-check # github.com/JonathonGore/api-check

# Run our unit tests
echo 'Running go unit tests...'
go test ./...

kill $PROJ_PID
