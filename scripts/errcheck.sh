#!/usr/bin/env bash

if ! which errcheck > /dev/null; then
    echo "==> Installing errcheck..."
    go install github.com/kisielk/errcheck@latest
fi

err_files=$(
    # shellcheck disable=SC2046
    errcheck \
        -ignoretests \
        -ignore 'bytes:.*' \
        -ignore 'io:Close|Write' \
    $(go list ./...| grep -v /vendor/) \
)

if [[ -n ${err_files} ]]; then
    echo "Unchecked errors found in the following places:"
    echo "${err_files}"
    echo "Please handle returned errors."
    exit 1
fi

exit 0
