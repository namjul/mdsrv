#!/bin/sh

set -e

version="$(git describe --tags)"
build="$(git log -n 1 --format='format: +%h %cd' HEAD)"

tags="netgo"
ldflags=$(printf -- "-X 'main.version=%s' -X 'main.build=%s'" "$version" "$build")

[ ! -d bin ] && mkdir bin

set -x
go build -tags "$tags" -ldflags "$ldflags" -o bin/mdsrv
set +x
