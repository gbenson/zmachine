all: build

.PHONY: build check run

check:
	gofmt -w .
	go vet ./...

run: check
	go run ./cmd/zmachine
