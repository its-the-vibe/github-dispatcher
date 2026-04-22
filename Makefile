.PHONY: build test lint

build:
	go build -o github-dispatcher .

test:
	go test -short ./...

lint:
	go vet ./...
