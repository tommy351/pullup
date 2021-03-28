---
id: architecture
title: Architecture
---

Pullup is composed by two components, a webhook and a controller. The diagram below describes how Pullup works. The left half of the diagram is managed by the webhook, and the right half is managed by the controller.

![Architecture](/img/architecture.png)

## Webhook

Webhook receives incoming events from GitHub or HTTP, and executes matched `Trigger` resources. When a `Trigger` is triggered, it will create, update or delete `ResourceTemplate` resources, which will be consumed by Pullup controller later.

## Controller

Controller maintains resource states according to `ResourceTemplate` resources.

When a new `ResourceTemplate` is created, the controller will create new resources according to the templates, or copy existing resources and mutate copied resources with [Strategic Merge Patch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md) or [JSON Patch](http://jsonpatch.com/).

After resources are created, the controller will track updates of `ResourceTemplate`. When a `ResourceTemplate` is updated, the resources created by it will also be updated automatically.

Finally, when a `ResourceTemplate` is deleted, the resources created by it will be deleted by [Garbage Collection](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/).
