FROM golang:1.12-alpine AS base

RUN apk add --update --no-cache git ca-certificates

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

ENV CGO_ENABLED=0
COPY cmd cmd
COPY pkg pkg
RUN go build -o /usr/local/bin/pullup -tags netgo -ldflags "-w" ./cmd/pullup

CMD ["pullup"]
