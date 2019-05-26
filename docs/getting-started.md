# Getting Started

## Creating a Webhook

The following is an example of a webhook.

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
        selector:
          matchLabels:
            app: "{{ .Name }}"
        template:
          metadata:
            labels:
              app: "{{ .Name }}"
          spec:
            - name: foo
              image: "tommy351/foo:{{ .Spec.Head.SHA }}"
    - apiVersion: v1
      kind: Service
      metadata:
        name: example
      spec:
        selector:
          app: "{{ .Name }}"
```

You can check the webhook list after creating the webhook on Kubernetes.

```sh
kubectl get webhooks.pullup.dev
# Or webhook for short
kubectl get webhook
```

### Repositories

The `repositories` array indicates which repositories to handle.

```yaml
repositories:
  - type: github
    name: tommy351/pullup
```

### Resources

The `resources` array indicates resources to apply when a pull request is opened or updated. `apiVersion`, `kind` and `metadata.name` are required for a resource.

The webhook and the resources to apply must be in the same namespace. Because [Garbage Collection](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/) in Kubernetes disallows cross-namespace references.

```yaml
resources:
  - apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: example
    spec:
      selector:
        matchLabels:
          app: "{{ .Name }}"
      template:
        metadata:
          labels:
            app: "{{ .Name }}"
        spec:
          - name: foo
            image: "tommy351/foo:{{ .Spec.Head.SHA }}"
  - apiVersion: v1
    kind: Service
    metadata:
      name: example
    spec:
      selector:
        app: "{{ .Name }}"
```

When a pull request is opened or updated, Pullup will search the existing resource by `apiVersion`, `kind` and `metadata.name`, then merge the `resources` array into the existing resource, finally create resources using the merged result. If the resources does not exist before, Pullup will create resources using the `resources` array directly. See the table below for example.

<table>
<thead>
  <tr>
    <th>Original</th>
    <th>Patched</th>
  </tr>
</thead>
<tbody>
  <tr>
    <td>
      <pre>apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
        - name: foo
          image: tommy351/foo</pre>
    </td>
    <td>
      <pre>apiVersion: apps/v1
kind: Deployment
metadata:
  name: <strong>example-123</strong>
spec:
  replicas: 1
  selector:
    matchLabels:
      app: <strong>example-123</strong>
  template:
    metadata:
      labels:
        app: <strong>example-123</strong>
    spec:
      containers:
        - name: foo
          image: <strong>tommy351/foo:dc970c24bbd20df017be64e110511d416eeddb36</strong></pre>
    </td>
  </tr>
  <tr>
    <td>
      <pre>apiVersion: v1
kind: Service
metadata:
  name: example
spec:
  selector:
    app: example
  ports:
    - port: 80</pre>
    </td>
    <td>
      <pre>apiVersion: v1
kind: Service
metadata:
  name: <strong>example-123</strong>
spec:
  selector:
    app: <strong>example-123</strong>
  ports:
    - port: 80</pre>
    </td>
  </tr>
</tbody>
</table>

## Registering on GitHub

After creating the webhook on your Kubernetes cluster. You can register it on your GitHub  repository or organization.

- **Payload URL**: `http://your-site.com/webhooks/github`
- **Content Type**: `application/json`
- **Secret**: See [Securing Your Webhooks](#securing-your-webhooks) below.
- **Events**: Choose **Pull Requests** only.

More details: https://developer.github.com/webhooks/creating/

## Securing Your Webhooks

It is recommended to set a secret on your webhook in order to make sure the payload is sent from GitHub. You can enable it by running `pullup-webhook` with `GITHUB_SECRET` environment variable. For example:

```yaml
env:
  - name: GITHUB_SECRET
    valueFrom:
      secretKeyRef:
        key: github-secret
        name: pullup
```
