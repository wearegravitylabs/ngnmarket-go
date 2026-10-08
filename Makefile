.PHONY: test vet fmt quality integration examples

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

# fmt, tidy, regenerate mocks, vet, test and lint.
quality:
	./quality.sh

# Live API calls (about six). Requires NGNMARKET_API_KEY.
integration:
	go test -tags integration -run Live -count=1 -v ./api/

# Make sure every sample still compiles.
examples:
	go build ./examples/...
