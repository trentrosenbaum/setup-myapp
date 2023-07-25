.PHONY: build test clean

BINARY_NAME := setup-myapp

build:
	go build -o $(BINARY_NAME)

test:
	go test

clean:
	go clean
	rm -f $(BINARY_NAME)
