.PHONY: build install test vet fmt tidy clean

BINARY := terraform-provider-auth0
GOBIN  := $(shell go env GOPATH)/bin

build:
	go build -o $(BINARY) .

install:
	go install .
	@echo "Installed $(BINARY) to $(GOBIN)"

test:
	go test ./... -count=1

vet:
	go vet ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
	go clean -cache

