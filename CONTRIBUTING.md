# Contributing

## Prerequisites

Pullup requires Go 1.12 and later.

## Getting Started

Download dependencies.

```sh
go get ./...
```

## Code Generation

If you modify `pkg/apis` or wire-related files, you have to regenerate code.

```sh
./hack/update-codegen.sh
```

## Testing

Download testing assets.

```sh
./hack/download-test-assets.sh
```

Run tests.

```sh
./hack/run-tests.sh
```

## Linting

Install [golangci-lint](https://github.com/golangci/golangci-lint) and run.

```sh
golangci-lint run
```
