# Troubleshooting

## Check if the webhook is triggered by pull requests

Try list events using `kubectl describe`.

```sh
kubectl describe webhook <webhook-name>
```

Or see if corresponding resource sets are created.

```sh
kubectl get resourceset
```

## Resources are not created

List events using `kubectl describe`.

```sh
kubectl describe resourceset <resourceset-name>
```
