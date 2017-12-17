.PHONY: test

test:
	go test ./...

build:
	go build -o bin/fargate main.go
