#!/bin/bash

KB_HOST=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | cut -d\  -f2)
psql -U user -d kbase -h ${KB_HOST}
