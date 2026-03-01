all: build

.PHONY: build check lint test run

check: test

lint:
	gofmt -w .
	go vet ./...

test: lint
	go test -coverprofile=coverage.out $(shell bash list-test-dirs.sh)

run: check
	go run ./cmd/zmachine

coverage.html: coverage.out
	go tool cover -html=$< -o $@
