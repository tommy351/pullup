# Pullup

[![GitHub release](https://img.shields.io/github/release/tommy351/pullup.svg)](https://github.com/tommy351/pullup/releases) [![CircleCI](https://circleci.com/gh/tommy351/pullup/tree/master.svg?style=svg)](https://circleci.com/gh/tommy351/pullup/tree/master) [![codecov](https://codecov.io/gh/tommy351/pullup/branch/master/graph/badge.svg)](https://codecov.io/gh/tommy351/pullup)

Pullup is a Kubernetes add-on that helps you deploy pull requests on a Kubernetes cluster based on existing resources and cleanup resources automatically when pull requests are closed or merged.

## Prerequisites

Pullup requires Kubernetes 1.7 and later which supports Custom Resource Definitions (CRD).

## Installation

Install Pullup CRDs and components in `pullup` namespace.

```sh
# Install the latest version
kubectl apply -f https://github.com/tommy351/pullup/releases/latest/download/pullup-deployment.yml

# Install a specific version
kubectl apply -f https://github.com/tommy351/pullup/releases/download/v0.3.3/pullup-deployment.yml
```

The YAML file is generated with [kustomize](https://github.com/kubernetes-sigs/kustomize). You can see source files in [deployment](deployment) folder. It contains:

- Pullup custom resource definitions (CRD).
- A Service account.
- RBAC for accessing Pullup CRD, writing events and the leader election.
- Deployments of the controller and the webhook.
- A service exposing the webhook server.

## RBAC

After Pullup is installed, you have to grant access of the resources that you defined in webhooks. If your Kubernetes cluster is not RBAC enabled, you can skip this section.

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
