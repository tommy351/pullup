---
id: troubleshooting
title: Troubleshooting
---

## Check if the webhook is triggered

List webhook events using `kubectl describe`.

```bash
kubectl describe httpwebhook <webhook-name>
```

If the webhook is triggered successfully, you should see events as below.

```
Events:
  Type    Reason     Age                From            Message
  ----    ------     ----               ----            -------
  Normal  Triggered  4s (x9 over 5d8h)  pullup-webhook  Triggered: default/kuard
```

## Check if the trigger is executed

List trigger events using `kubectl describe`.

```bash
kubectl describe trigger <trigger-name>
```

If the trigger is executed successfully, you should see events as below.

```
Events:
  Type    Reason   Age    From            Message
  ----    ------   ----   ----            -------
  Normal  Updated  2m17s  pullup-webhook  Updated resource template: kuard-test
```

## Check if the resource template is updated

Check the last update time of resource templates.

```bash
kubectl get resourcetemplate <name>
```

if the resource template is updated, the last update time should be updated.

```
NAME         TRIGGER   LAST UPDATE   AGE
kuard-test   kuard     6s            6s
```

You can check resource template events for detailed information.

```bash
kubectl describe resourcetemplate <name>
```
