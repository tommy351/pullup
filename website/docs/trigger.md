---
id: trigger
title: Trigger
---

`Trigger` defines resources to create and an optional JSON schema of input data. When a `Trigger` is executed, Pullup controller will create, update or delete resources according to input data.

## Examples

### Create Resources from Scratch

This is the most basic usage of the `Trigger` resource.

```yaml
apiVersion: pullup.dev/v1beta1
kind: Trigger
metadata:
  name: example
spec:
  resourceName: "{{ .trigger.metadata.name }}"
  patches:
    - apiVersion: v1
      kind: ConfigMap
      merge:
        data:
          timezone: UTC
```

When the trigger above is executed, a `ConfigMap` will be created as below.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example
data:
  timezone: UTC
```

### Copy and Mutate Resources

### Validate Input Data

### Customize Resource Name

## References
