#!/bin/bash

KBASE_UI=/Users/jack/Projects/knowledge-base-ui

# Build docker images
docker build --no-cache -t jackgore/knowledge-base .
docker build --no-cache -f ${KBASE_UI}/Dockerfile.local -t jackgore/knowledge-base-ui ${KBASE_UI}

# Push the docker images
docker push jackgore/knowledge-base
docker push jackgore/knowledge-base-ui

ssh kbase '/bin/bash -s' < scripts/remote-restart.sh
