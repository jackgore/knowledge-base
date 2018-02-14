#!/bin/bash

cert_name=server.crt
key_name=server.key

# Displays usage to the user
usage () {
	echo "usage: generateCerts.sh <import path>"
	echo "	generates self signed certificate and key for ssl server"
	echo "---------------------------------"
	echo "Example:"
	echo "	generateCerts.sh --cert-name=go-service --key-name=go-service"
	echo "Note: the flags are optional and will default to 'server'"
	exit
}

# Handle various options and flags
while [ "$#" -gt 0 ]; do
  case "$1" in
    -h) usage; shift 2;;
    --help) usage; shift 2;;

	--cert-name=*) cert_name="${1#*=}"; shift 1;;
	--key-name=*) key_name="${1#*=}"; shift 1;;

    -*) echo "unknown option: $1" >&2; exit 1;;
  esac
done

echo 'Generating self signed server.key and server.crt file'
openssl genrsa -out ${key_name} 2048
openssl req -new -x509 -sha256 -key ${key_name} -out ${cert_name} -days 3650
