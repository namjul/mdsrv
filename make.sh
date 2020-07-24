#!/bin/sh

set -e

BUILD="-X=main.build=$(git rev-parse HEAD) -X=main.version=$(git tag | head -n 1)"

set -x
go build -tags "netgo" -ldflags "$BUILD" -o mdsrv.out
set +x
