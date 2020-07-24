#!/bin/sh

set -e

BUILD="-X=main.build=$(git rev-parse HEAD)"

set -x
go build -tags "netgo" -ldflags "$BUILD" -o mdsrv.out
set +x
