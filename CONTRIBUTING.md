# Contributing

## Prerequisites

Pullup requires Go 1.15 and later.

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

### Unit Tests

Download testing assets.

```sh
./hack/download-test-assets.sh
```

Run tests.

```sh
./hack/run-tests.sh
```

### End-to-end Tests

Before you get started, make sure your environment can run [kind].

Build Docker images.

```sh
./test/scripts/build-images.sh
```

Run tests. This script will setup a Kubernetes cluster using [kind] and run the tests. After all tests are done, the cluster will be teardown automatically.

```sh
./test/scripts/kind/run.sh
```

## Linting

Install [golangci-lint](https://github.com/golangci/golangci-lint) and run.

```sh
golangci-lint run
```

[kind]: https://kind.sigs.k8s.io/
