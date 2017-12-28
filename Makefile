.PHONY: test build dist

test:
	go test ./...

build:
	go get github.com/spf13/cobra
	go get github.com/aws/aws-sdk-go/aws
	go get github.com/mgutz/ansi
	go get github.com/hashicorp/golang-lru
	go get golang.org/x/time/rate
	go get golang.org/x/crypto/ssh/terminal

	go build -o bin/fargate main.go

dist:
	GOOS=darwin GOARCH=amd64 go build -o dist/build/fargate-darwin-amd64/fargate main.go
	GOOS=linux GOARCH=amd64 go build -o dist/build/fargate-linux-amd64/fargate main.go
	GOOS=linux GOARCH=386 go build -o dist/build/fargate-linux-386/fargate main.go
	GOOS=linux GOARCH=arm go build -o dist/build/fargate-linux-arm/fargate main.go

	cd dist/build/fargate-darwin-amd64 && zip fargate-${FARGATE_VERSION}-darwin-amd64.zip fargate
	cd dist/build/fargate-linux-amd64 && zip fargate-${FARGATE_VERSION}-linux-amd64.zip fargate
	cd dist/build/fargate-linux-386  && zip fargate-${FARGATE_VERSION}-linux-386.zip fargate
	cd dist/build/fargate-linux-arm  && zip fargate-${FARGATE_VERSION}-linux-arm.zip fargate

	find dist/build -name *.zip -exec mv {} dist \;

	rm -rf dist/build
