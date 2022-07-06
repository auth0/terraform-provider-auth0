#!/usr/bin/env bash

set -eu -o pipefail

if ! command -v golangci-lint &> /dev/null ; then
    echo "==> Installing golangci-lint" >&2
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
fi

exec golangci-lint run -c .golangci.yml --allow-parallel-runners --fix ./...
