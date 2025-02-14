#!/bin/bash
set -e

if ! command -v mockgen &>/dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v mockgen &>/dev/null; then
        echo "installing mockgen..."
        go install go.uber.org/mock/mockgen@latest
    fi
fi

go generate .
