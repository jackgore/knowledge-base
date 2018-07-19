#!/bin/bash

KBASE_UI=${PROJECTS_HOME}/knowledge-base-ui
KBASE=${GOPATH}/src/github.com/JonathonGore/knowledge-base

# Build docker images
docker build --no-cache -f ${KBASE}/Dockerfile -t jackgore/knowledge-base ${KBASE}
docker build --no-cache -f ${KBASE_UI}/Dockerfile.local -t jackgore/knowledge-base-ui ${KBASE_UI}

# Push the docker images
docker push jackgore/knowledge-base
docker push jackgore/knowledge-base-ui

ssh kbase '/bin/bash -s' < scripts/remote-restart.sh
