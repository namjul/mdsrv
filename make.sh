#!/bin/sh

set -e

version="$(git rev-parse --abbrev-ref HEAD)"

if ! echo "$version" | grep -qE "^v"; then
	version="devel $(git log -n 1 --format='format: +%h %cd' HEAD)"
fi

tags="netgo"
ldflags=$(printf -- "-X 'main.version=%s'" "$module" "$version")

[ ! -d bin ] && mkdir bin

set -x
go build -tags "$tags" -ldflags "$ldflags" -o bin/mdsrv
set +x
