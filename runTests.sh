#!/bin/bash

# Install our project and output any errors
go install

# Run our server
knowledge-base > test_logs.txt 2>&1 &

PROJ_PID=$!

# Run our cucumber tests
cucumber

kill $PROJ_PID
