# pullup

[![GitHub release](https://img.shields.io/github/release/tommy351/pullup.svg)](https://github.com/tommy351/pullup/releases) [![CircleCI](https://circleci.com/gh/tommy351/pullup/tree/master.svg?style=svg)](https://circleci.com/gh/tommy351/pullup/tree/master) [![codecov](https://codecov.io/gh/tommy351/pullup/branch/master/graph/badge.svg)](https://codecov.io/gh/tommy351/pullup)

Deploy pull requests on a Kubernetes cluster before merged.

## Installation

Install CRDs.

```sh
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/crds/webhook.yml
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/crds/resource-set.yml
```

Create a new service account and RBAC. By default, it only grants access to deployments, services and pullup resources. You may need to modify RBAC if you want to deploy more kinds of resources.

```sh
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/rbac.yml
```

Create a new deployment. It runs a controller that monitoring resource changes and starts a HTTP server listening on port 4000.

```sh
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/deployment.yml
```

Create a new service.

```sh
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/service.yml
```

## Definitions

### Webhook

```yaml
apiVersion: pullup.dev/v1alpha1
kind: Webhook
metadata:
  name: example
spec:
  github:
    secret: ''
  # Resources to apply when receiving GitHub webhooks
  resources:
    - apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: example
      spec:
        template:
          spec:
            - name: foo
              image: tommy351/foo:{{ .Spec.Head.SHA }}
    - apiVersion: v1
      kind: Service
      metadata:
        name: example
```

## Configuration

Flag | Description | Environment Variable | Default
--- | --- | --- | ---
`log.level` | Log level. Possible values: `debug`, `info`, `warn`, `error`, `panic`, `fatal` | `LOG_LEVEL` | `info`
`namespace` | Kubernetes namespace | `KUBERNETES_NAMESPACE` | `default`
`kubeconfig` | Kubernetes config path. Don't set this unless it's run out of cluster. | `KUBECONFIG` |
`webhook-address` | Webhook listening address | `WEBHOOK_ADDRESS` | `:4000`

## Development

Run tests.

```sh
./hack/run-test.sh
```

Run integration tests.

```sh
./hack/download-test-assets.sh
./hack/run-test.sh -tags integration
```

Generate code.

```sh
./hack/update-codegen.sh
```

## License

MIT
