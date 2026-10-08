#!/bin/sh
set -e

echo "running go fmt"
go fmt ./...
echo "running go mod tidy"
go mod tidy
echo "running go generate (regenerates api/mock)"
go generate ./...
echo "running go vet"
go vet ./...
echo "running tests"
go test -race -count=1 ./...
echo "running golangci-lint (v2 required)"
# Prefer the copy installed with `go install` so an older golangci-lint earlier
# on PATH (e.g. in /usr/local/bin) is not picked up.
LINT="$(go env GOPATH)/bin/golangci-lint"
[ -x "$LINT" ] || LINT=golangci-lint
"$LINT" run ./... -v --config .golangci.yml
echo "quality check done"
