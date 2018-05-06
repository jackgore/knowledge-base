#!/usr/local/bin/python

import requests
import sys

args = sys.argv

if len(args) != 2:
	print("usage: getquestion.py <question-id>")
	sys.exit(0)

question = int(args[1])

r = requests.get("http://localhost:3001/questions/" + str(question))

print("received status code: ", r.status_code)
print("response body: ", r.text)
