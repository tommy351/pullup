FROM golang:1.12 AS base

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

ENV CGO_ENABLED=0
COPY cmd cmd
COPY pkg pkg
RUN go build -o /usr/local/bin/pullup -tags netgo -ldflags "-w" ./cmd/pullup

FROM gcr.io/distroless/base
COPY --from=base /usr/local/bin /usr/local/bin
CMD ["pullup"]
