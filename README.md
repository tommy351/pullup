# pullup

[![GitHub release](https://img.shields.io/github/release/tommy351/pullup.svg)](https://github.com/tommy351/pullup/releases) [![CircleCI](https://circleci.com/gh/tommy351/pullup/tree/master.svg?style=svg)](https://circleci.com/gh/tommy351/pullup/tree/master) [![codecov](https://codecov.io/gh/tommy351/pullup/branch/master/graph/badge.svg)](https://codecov.io/gh/tommy351/pullup)

Deploy pull requests on a Kubernetes cluster before merged.

## Features

- Create new resources based on existing resources when pull requests are opened.
- Cleanup resources automatically when pull requests are closed.

## Documentation

- [Getting Started](docs/getting-started.md)
- [References](docs/references.md)
- [Troubleshooting](docs/troubleshooting.md)

## Development

Run tests.

```sh
./hack/download-test-assets.sh
./hack/run-test.sh
```

Generate code.

```sh
./hack/update-codegen.sh
```

## Todos

- [ ] Merge resources using [Structured Merge and Diff](https://github.com/kubernetes-sigs/structured-merge-diff)
- [ ] Test more kinds of resources
