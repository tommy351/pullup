# Getting Started

## Installation

Pullup requires Kubernetes 1.7 and above which supports Custom Resource Definitions (CRD).

First, create a new namespace.

```sh
kubectl create namespace pullup
```

Install CRDs.

```sh
kubectl apply -f https://github.com/tommy351/pullup/blob/master/deployment/crds/webhook.yml
kubectl apply -f https://github.com/tommy351/pullup/blob/master/deployment/crds/resource-set.yml
```

Create a new service account and RBAC if it is enabled on your cluster. This enables pullup to access pullup custom resources.

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

## First Webhook

### Creating a Webhook

Let's start by creating the very first webhook. The following is a basic example of a webhook.

```yaml
apiVersion: pullup.dev/v1alpha1
kind: Webhook
metadata:
  name: example
spec:
  repositories:
    - type: github
      name: tommy351/pullup
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
```

The `repositories` array indicates which repositories to handle.

```yaml
spec:
  repositories:
    - type: github
      name: tommy351/pullup
```

The `resources` array indicates resources to apply when a pull request is opened or updated. `apiVersion`, `kind` and `metadata.name` are required for a resource. Pullup will try to find the resource with the given name, otherwise a new one will be created.

The webhook and the resources to apply must be in the same namespace. Because [Garbage Collection](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/) in Kubernetes disallows cross-namespace references.

```yaml
spec:
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
```

### Registering on GitHub

After creating the webhook on your Kubernetes cluster. You can register it on your GitHub organization or repository.

- **Payload URL**: `http://your-site.com/webhooks/github`
- **Content Type**: `application/json`
- **Secret**: See [Securing Your Webhooks](#securing-your-webhooks) below.
- **Events**: Choose **Pull Requests** only

More details: https://developer.github.com/webhooks/creating/

### Securing Your Webhooks

It is recommended to set a secret on your webhook in order to make sure the payload is sent from GitHub. You can enable it by running `pullup-webhook` with `GITHUB_SECRET` environment variable. For example:

```yaml
env:
  - name: GITHUB_SECRET
    valueFrom:
      secretKeyRef:
        key: github-secret
        name: pullup
```

### Opening a Pull Request

When a pull request is opened, GitHub will send a event to pullup webhook. Pullup will find a matching webhook based on the incoming event and create or update a resource set.

Pullup controller monitors changes of resource sets. When a resource set is created or updated, the controller will create resources defined in the resource set.

### Closing a Pull Request

When a pull request is closed or merged. Pullup will delete matching resource sets.
