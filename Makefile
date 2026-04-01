all: build

.PHONY: build check lint test run install

build: zmachine

check: test

lint:
	gofmt -w .
	go vet ./...

test: lint
	go test -coverprofile=coverage.out $(shell bash list-test-dirs.sh)

run: check
	go run ./cmd/zmachine

zmachine: check
	go build -o $@ ./cmd/zmachine

install:
	@bash escape-sudo.sh $(MAKE) zmachine
	install -m755 zmachine /usr/bin

coverage.html: coverage.out
	go tool cover -html=$< -o $@
