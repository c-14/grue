.PHONY: all build test

all: build vet fmt

build: grue.go
	go build

fmt: grue.go
	go fmt

vet: grue.go
	go vet

# test: grue_test.go
# 	go test
