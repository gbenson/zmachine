GOLANG_VERSION ?= $(shell sed -n 's/^go //p' go.mod)
export BUILDER_IMAGE ?= golang:$(GOLANG_VERSION)-trixie
export BUILDER_UID ?= $(shell id -u)
export BUILDER_GID ?= $(shell id -g)

all: build

.PHONY: build builder check lint test run install

build: zmachine

builder:
	docker compose run --rm -it builder

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
