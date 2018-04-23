#!/usr/local/bin/python

import requests
import sys

args = sys.argv

if len(args) != 3:
	print("usage: login.py <username> <password>")
	sys.exit(0)

username = args[1]
password = args[2]

r = requests.post("http://localhost:3001/login", json={'username': username, 'password': password})

print("received status code ", r.status_code)
print("response body ", str(r.content))
