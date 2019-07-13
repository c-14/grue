.PHONY: all build test

GO111MODULES=on
export GO111MODULES

all: build vet fmt

build: grue.go
	go build

fmt: grue.go
	go fmt

vet: grue.go
	go vet

# test: grue_test.go
# 	go test
