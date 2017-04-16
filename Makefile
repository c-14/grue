.PHONY: all build test

all: build test

build: grue.go
	go build

test: grue_test.go
	go test
