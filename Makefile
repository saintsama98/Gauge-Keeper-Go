.PHONY: build run tidy fmt vet test clean

build:
	go build -o bin/keeper ./cmd/keeper

run:
	go run ./cmd/keeper

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf bin
