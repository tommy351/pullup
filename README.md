# pullup

[![GitHub release](https://img.shields.io/github/release/tommy351/pullup.svg)](https://github.com/tommy351/pullup/releases) [![CircleCI](https://circleci.com/gh/tommy351/pullup/tree/master.svg?style=svg)](https://circleci.com/gh/tommy351/pullup/tree/master) [![codecov](https://codecov.io/gh/tommy351/pullup/branch/master/graph/badge.svg)](https://codecov.io/gh/tommy351/pullup)

Pullup is a Kubernetes add-on that helps you deploy pull requests on a Kubernetes cluster based on existing resources and cleanup resources automatically when pull requests are closed or merged.

## Prerequisites

Pullup requires Kubernetes 1.7 and later which supports Custom Resource Definitions (CRD).

## Installation

First, create a new namespace.

```sh
kubectl create namespace pullup
```

Install CRDs.

```sh
kubectl apply -f https://github.com/tommy351/pullup/blob/master/deployment/crds/webhook.yml
kubectl apply -f https://github.com/tommy351/pullup/blob/master/deployment/crds/resource-set.yml
```

Create a new service account and RBAC if it is enabled on your Kubernetes cluster. This enables Pullup to access Pullup resources and the leader election.

```sh
kubectl apply -f https://github.com/tommy351/pullup/blob/master/deployment/rbac.yml
```

Create deployments. This will create two deployments. One is the controller that monitoring resource changes and the other is a HTTP server receiving GitHub events.

```sh
kubectl apply -f https://github.com/tommy351/pullup/blob/master/deployment/deployment.yml
```

Create a new service exposing the webhook server. You may need to change the service type based on your need.

```sh
kubectl apply -f https://github.com/tommy351/pullup/blob/master/deployment/service.yml
```

## RBAC

Besides the RBAC settings you have installed in the previous section, you also have to grant access of the resources that you defined in webhooks. If your Kubernetes cluster is not RBAC enabled, you can skip this section.

You have to create `Role` and `RoleBinding` (or `ClusterRole` and `ClusterRoleBinding` for all namespaces), set `verbs` to `["get", "create", "update"]` for each kind of resources and bind the role to the pullup service account.

The following example includes `Deployment` and `Service`. See [here](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) for more details about RBAC.

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
# Change this to ClusterRole to apply in all namespaces.
kind: Role
metadata:
  name: pullup-deployment
rules:
  # Deployment
  - apiGroups: ["apps", "extensions"]
    resources: ["deployments"]
    verbs: ["get", "create", "update"]
  # Service
  - apiGroups: [""]
    resources: ["services"]
    verbs: ["get", "create", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
# Change this to ClusterRoleBinding to apply in all namespaces.
kind: RoleBinding
metadata:
  name: pullup-deployment
# Bind to the pullup service account
subjects:
  - kind: ServiceAccount
    name: pullup
    namespace: pullup
# Refer to the role above
roleRef:
  kind: Role
  name: pullup-deployment
  apiGroup: rbac.authorization.k8s.io
```

## Documentation

- [Getting Started](docs/getting-started.md)
- [Architecture](docs/architecture.md)
- [References](docs/references.md)
- [Troubleshooting](docs/troubleshooting.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for information about setting environment.

## Todos

- [ ] Merge resources using [Structured Merge and Diff](https://github.com/kubernetes-sigs/structured-merge-diff)
- [ ] Test more kinds of resources
- [ ] End-to-end tests
