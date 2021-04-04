---
id: architecture
title: Architecture
---

Pullup consists of two components: webhook and controller.

## Webhook

The webhook listens to GitHub events.

When the webhook receive a event, it will lookup webhooks by the `repositories` array.

- When a pull request is **opened** or **reopened**, the webhook will create a new `ResourceSet` based on `Webhook`.
- When a pull request is **updated**, the webhook will update the existing `ResourceSet`.
- When a pull request is **closed** or merged, the webhook will delete `ResourceSet` matching the labels `webhook-name` and `pull-request-number`.

## Controller

The controller monitors changes of `Webhook` and `ResourceSet`.

### Webhook

- When a `Webhook` is updated, the controller will update `spec.resources` in `ResourceSet` matching the label `webhook-name`.

### ResourceSet

- When a `ResourceSet` is **created**, the controller will find the original resources by `apiVersion`, `kind` and `metadata.name`.
  - If the resources **exist**, the controller will merge the `resources` array into the original resources and create resources using the merged result.
  - If the resources does **not exist**, the controller will create resources using the `resources` array directly.
- When a `ResourceSet` is **updated**, the controller will also merge the current resource before merging the `resources` array.

### Merge Strategy

- **Map**
  - Replace existing keys with new values.
  - Copy new keys.
- **Array**
  - If all elements in the array contain `name` key, the array will be merged using the `name` key.
  - Otherwise, replace with new values.

[Structured Merge and Diff](https://github.com/kubernetes-sigs/structured-merge-diff) may be implemented in the future.
