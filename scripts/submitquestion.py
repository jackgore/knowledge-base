#!/usr/local/bin/python

import requests
import sys

args = sys.argv

if len(args) < 4:
	print("usage: submitquestion.py <author-id> <title> <content> [<team> <org>]")
	sys.exit(0)

author = int(args[1])
title = args[2]
content = args[3]

file = open(".session", "r") # We need a valid session ID to be in the .session file
cookies = {'knowledge_base': (file.read())[:-1]} # Remove newline from cookie

if len(args) == 6:
	team = args[4]
	org = args[5]
	r = requests.post(f"http://localhost:3001/organizations/{org}/teams/{team}/questions", 
			cookies=cookies, json={'author': author, 'title': title, 'content': content})
else:
	r = requests.post("http://localhost:3001/questions", json={'author': author, 'title': title, 'content': content}, cookies=cookies)


print("received status code: ", r.status_code)
print("response body: ", r.text)
