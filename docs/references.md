# References

## Webhook

Webhook defines repositories to handle and resources to apply.

```yaml
apiVersion: pullup.dev/v1alpha1
kind: Webhook
metadata:
  name: example
spec:
  repositories:
    - type: github
      name: tommy351/pullup
  resourceName: "{{ .Webhook.Name }}-{{ .Spec.Number }}"
  resources:
    - apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: example
      spec:
        template:
          spec:
            - name: foo
              image: "tommy351/pullup:{{ .Spec.Head.SHA }}"
    - apiVersion: v1
      kind: Service
      metadata:
        name: example
```

Name | Type | Description | Default
--- | --- | --- | ---
`spec.repositories` | `array` | Repositories to handle. |
`spec.repositories[].type` | `string` | Repository type. Possible values: `github` | **Required**
`spec.repositories[].name` | `string` | Full name of repository | **Required**
`spec.repositories[].branches.include` | `[]string` | Included branches. |
`spec.repositories[].branches.exclude` | `[]string` | Excluded branches. |
`spec.resourceName` | `string` | Template of resource name. This value is used to generate name of resource sets. | `{{ .Webhook.Name }}-{{ .Spec.Number }}`
`resources` | `array` | Resources to apply. |

### Filter by Branch

You can filter pull requests by setting `spec.repositories[].branches`. The value can be a string or a regular expression. For example:

```yaml
- master
- /.*/
```

### Template

You can use template strings in `spec.resourceName` or `spec.resources`.

Webhook variables:

- `.Spec`: [ResourceSetSpec](https://godoc.org/github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1#ResourceSetSpec)
- `.Webhook`: [Webhook](https://godoc.org/github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1#Webhook)
- `.Repo`: [Repository](https://godoc.org/github.com/google/go-github/github#Repository)

ResourceSet variables:

- `.`: [ResourceSet](https://godoc.org/github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1#ResourceSet)

See more:

- Syntax: [text/template](https://golang.org/pkg/text/template/)
- Available functions: [sprig](https://masterminds.github.io/sprig/)

## ResourceSet

ResourceSet defines resources to apply and some information about the pull request and the commit. It's usually created by the webhook.

```yaml
apiVersion: pullup.dev/v1alpha1
kind: ResourceSet
metadata:
  name: example-123
spec:
  base:
    ref: master
    sha: "3afa0879385842fa7423a8f18ab03783709a3d3e"
  head:
    ref: feature
    sha: "121b29fb6c467d388faea3d7c9b4859b7e244772"
  number: 123
  resources: []
```

Name | Type | Description
--- | --- | ---
`spec.base.ref` | `string` | Base branch name.
`spec.base.sha` | `string` | Base branch revision.
`spec.head.ref` | `string` | Head branch name.
`spec.head.sha` | `string` | Head branch revision.
`spec.number` | `int` | Pull requset number.
`spec.resources` | `array` | Resources to apply.
