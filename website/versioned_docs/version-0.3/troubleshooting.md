---
id: troubleshooting
title: Troubleshooting
---

## Check if the webhook is triggered by pull requests

Try list events using `kubectl describe`.

```bash
kubectl describe webhook <webhook-name>
```

Or see if corresponding resource sets are created.

```bash
kubectl get resourceset
```

## Resources are not created

List events using `kubectl describe`.

```bash
kubectl describe resourceset <resourceset-name>
```
