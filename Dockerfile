FROM golang:alpine

RUN apk add --no-cache git upx

WORKDIR /fargate

ADD go.mod .
RUN go mod download

ADD . /fargate
RUN go build -ldflags="-s -w"
RUN upx --brute fargate

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=0 /fargate/fargate /usr/local/bin/
