PHONY: build run test

build:
	go build -o bin/server app/*.go

run: build
	./bin/server

test:
	go test -v ./...

clean:
	rm -rf bin
	go mod tidy
