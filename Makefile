.PHONY: test checks build

export GOOS=wasip1
export GOARCH=wasm

default: test checks build

test:
	go test -v -cover ./...

build:
	@go build -o plugin.wasm ./demo.go

checks:
	golangci-lint run
