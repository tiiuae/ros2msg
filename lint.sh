#!/bin/bash

set +x -eo pipefail 
if [[ $CI == true ]]; then
    env > "$GITHUB_ENV"
    set -x
    go vet -vettool=./checklocks.sh ./...
else
    set -x
    go vet -vettool=./checklocks.sh ./...
    golangci-lint run ./...
fi
