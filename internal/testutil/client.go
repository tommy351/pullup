package testutil

import (
	"context"

	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	sb := runtime.NewSchemeBuilder(v1alpha1.AddToScheme)

	if err := sb.AddToScheme(scheme); err != nil {
		panic(err)
	}

	return scheme
}

type nameGetter interface {
	GetNamespace() string
	GetName() string
}

type Client struct {
	client.Client

	Changed map[string]map[string]runtime.Object
	Deleted map[string]map[string]runtime.Object
}

func NewClient(objects ...runtime.Object) *Client {
	return &Client{
		Client:  fake.NewFakeClientWithScheme(NewScheme(), objects...),
		Changed: map[string]map[string]runtime.Object{},
		Deleted: map[string]map[string]runtime.Object{},
	}
}

func (c *Client) addToMap(m map[string]map[string]runtime.Object, obj runtime.Object) {
	gvk := obj.GetObjectKind().GroupVersionKind().String()

	if _, ok := c.Changed[gvk]; !ok {
		m[gvk] = map[string]runtime.Object{}
	}

	name := obj.(nameGetter)
	m[gvk][name.GetNamespace()+"/"+name.GetName()] = obj
}

func (c *Client) addChanged(obj runtime.Object) {
	c.addToMap(c.Changed, obj)
}

func (c *Client) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOptionFunc) error {
	c.addChanged(obj)
	return c.Client.Create(ctx, obj, opts...)
}

func (c *Client) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	c.addToMap(c.Deleted, obj)
	return c.Client.Delete(ctx, obj, opts...)
}

func (c *Client) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOptionFunc) error {
	c.addChanged(obj)
	return c.Client.Update(ctx, obj, opts...)
}

func (c *Client) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOptionFunc) error {
	c.addChanged(obj)
	return c.Client.Patch(ctx, obj, patch, opts...)
}
