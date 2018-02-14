#!/bin/bash

# The file in which we store the previous import path
RENAME_FILE=.go.service.rename

DEFAULT_IMPORT_PATH=github.com/JonathonGore/go-service

# Displays usage to the user
usage () {
	echo "usage: setup.sh <import path>"
	echo "---------------------------------"
	echo "Example:"
	echo "	setup.sh 'github.com/JonathonGore/my-service'"
}

assign_path () {
	IMPORT_PATH=$1
}

# Assert we have received a import path as an argument
if [ $# -eq 0 ]
  then
	usage
	exit
fi

# Handle various options and flags
while [ "$#" -gt 0 ]; do
  case "$1" in
    -*) echo "unknown option: $1" >&2; exit 1;;
    *) assign_path "$1"; shift 1;;
  esac
done

# Package required for renaming import paths
go get -u github.com/rogpeppe/govers

# Stores the new changed import path to allow it to be undone - TODO
# echo $1 > $RENAME_FILE

echo Renaming $DEFAULT_IMPORT_PATH to $IMPORT_PATH

# Rename import paths referencing 
govers -d -m $DEFAULT_IMPORT_PATH $IMPORT_PATH

echo Updating Dockerfile

# Note % is an alternate deliminater as opposed to / as our variables will likely have /'s
sed -i -e "s%${DEFAULT_IMPORT_PATH}%${IMPORT_PATH}%g" Dockerfile
