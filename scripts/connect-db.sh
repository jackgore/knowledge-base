#!/bin/bash

DB='kbase'

# Handle various options and flags
while [ "$#" -gt 0 ]; do
  case "$1" in
	--test) DB="test"; shift 1;;

    -*) echo "unknown option: $1" >&2; exit 1;;
  esac
done

KB_HOST=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | cut -d\  -f2)
PGPASSWORD=password psql -U kbase -d ${DB} -h ${KB_HOST}
