---
id: installation
title: Installation
slug: /
---

Install Pullup CRDs and components in `pullup` namespace.

```bash
# Install the latest version
kubectl apply -f https://github.com/tommy351/pullup/releases/latest/download/pullup-deployment.yml

# Install a specific version
kubectl apply -f https://github.com/tommy351/pullup/releases/download/v0.3.3/pullup-deployment.yml
```

The YAML file is generated with [kustomize](https://github.com/kubernetes-sigs/kustomize). You can see source files in [deployment](https://github.com/tommy351/pullup/tree/v0.3.6/deployment) folder. It contains:

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
