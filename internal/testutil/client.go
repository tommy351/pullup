package testutil

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type nameGetter interface {
	GetNamespace() string
	GetName() string
}

type Client struct {
	client.Client

	Changes map[string]map[string]runtime.Object
}

func (c *Client) addChange(obj runtime.Object) {
	if c.Changes == nil {
		c.Changes = map[string]map[string]runtime.Object{}
	}

	gvk := obj.GetObjectKind().GroupVersionKind().String()

	if _, ok := c.Changes[gvk]; !ok {
		c.Changes[gvk] = map[string]runtime.Object{}
	}

	name := obj.(nameGetter)
	c.Changes[gvk][name.GetNamespace()+"/"+name.GetName()] = obj
}

func (c *Client) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOptionFunc) error {
	c.addChange(obj)
	return c.Client.Create(ctx, obj, opts...)
}

func (c *Client) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	c.addChange(obj)
	return c.Client.Delete(ctx, obj, opts...)
}

func (c *Client) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOptionFunc) error {
	c.addChange(obj)
	return c.Client.Update(ctx, obj, opts...)
}

func (c *Client) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOptionFunc) error {
	c.addChange(obj)
	return c.Client.Patch(ctx, obj, patch, opts...)
}
