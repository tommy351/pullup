# pullup

[![GitHub release](https://img.shields.io/github/release/tommy351/pullup.svg)](https://github.com/tommy351/pullup/releases) [![CircleCI](https://circleci.com/gh/tommy351/pullup/tree/master.svg?style=svg)](https://circleci.com/gh/tommy351/pullup/tree/master) [![codecov](https://codecov.io/gh/tommy351/pullup/branch/master/graph/badge.svg)](https://codecov.io/gh/tommy351/pullup)

Deploy pull requests on a Kubernetes cluster before merged.

## Installation

Create a new namespace.

```sh
kubectl create namespace pullup
```

Install CRDs.

```sh
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/crds/webhook.yml
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/crds/resource-set.yml
```

Create a new service account and RBAC.

```sh
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/rbac.yml
```

Create a new deployment. It runs a controller that monitoring resource changes and starts a HTTP server listening on port 8080.

```sh
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/deployment.yml
```

Create a new service.

```sh
kubectl create -f https://github.com/tommy351/pullup/blob/master/deployment/service.yml
```

## Usage

List webhooks.

```sh
kubectl get webhook
```

List resource sets.

```sh
kubectl get resourceset
```

## Definitions

### Webhook

Webhook defines repositories to listen and resources to apply when a pull request is opened.

```yaml
apiVersion: pullup.dev/v1alpha1
kind: Webhook
metadata:
  name: example
spec:
  repositories:
    - type: github
      name: tommy351/pullup
  resourceName: "{{ .Webhook.Name }}-{{ .Spec.Number }}"
  resources:
    - apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: example
      spec:
        template:
          spec:
            - name: foo
              image: "tommy351/foo:{{ .Spec.Head.SHA }}"
    - apiVersion: v1
      kind: Service
      metadata:
        name: example
```

### ResourceSet

ResourceSet defines resources to apply and some information about the pull request and the commit. It's usually created by the webhook.

```yaml
apiVersion: pullup.dev/v1alpha1
kind: ResourceSet
metadata:
  name: example-123
spec:
  base:
    ref: master
    sha: 3afa0879385842fa7423a8f18ab03783709a3d3e
  head:
    ref: feature
    sha: 121b29fb6c467d388faea3d7c9b4859b7e244772
  number: 123
  resources: []
```

## Configuration

Flag | Description | Environment Variable | Default
--- | --- | --- | ---
`log.level` | Log level. Possible values: `debug`, `info`, `warn`, `error`, `panic`, `fatal` | `LOG_LEVEL` | `info`
`namespace` | Kubernetes namespace. This option is used for leader election of the controller. | `KUBERNETES_NAMESPACE` | `default`
`kubeconfig` | Kubernetes config path. Don't set this unless it's run out of cluster. | `KUBECONFIG` |
`address` | Webhook listening address | `WEBHOOK_ADDRESS` | `:8080`
`github-secret` | GitHub secret (See [Securing your webhooks](https://developer.github.com/webhooks/securing/) for more) | `GITHUB_SECRET` |

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
