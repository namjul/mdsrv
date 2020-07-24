#!/bin/sh

set -e

BUILD="-X=main.build=$(git rev-parse HEAD) -X=main.version=$(git describe --tags --abbrev=0)"

set -x
go build -tags "netgo" -ldflags "$BUILD" -o mdsrv.out
set +x
