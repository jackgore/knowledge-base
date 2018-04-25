#!/usr/local/bin/python

import requests
import sys

args = sys.argv

if len(args) != 4:
	print("usage: submitquestion.py <author-id> <title> <content>")
	sys.exit(0)

author = int(args[1])
title = args[2]
content = args[3]

r = requests.post("http://localhost:3001/questions", json={'author': author, 'title': title, 'content': content})

print("received status code: ", r.status_code)
print("response body: ", r.text)
