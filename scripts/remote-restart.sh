#!/bin/bash

sudo su
docker pull jackgore/knowledge-base
docker pull jackgore/knowledge-base-ui

/usr/local/bin/docker-compose down
/usr/local/bin/docker-compose up -d
