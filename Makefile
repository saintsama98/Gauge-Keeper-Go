.PHONY: build run tidy fmt vet test clean bindings

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

# Regenerate Go contract bindings from raw ABIs in ./abi (requires abigen on PATH).
# One invocation per contract — see abi/README.md for the file -> package mapping.
bindings:
	@echo "Run abigen per abi/README.md -> internal/bindings/<name>/<name>.gen.go"
