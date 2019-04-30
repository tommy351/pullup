package testutil

import (
	"context"

	"github.com/tommy351/pullup/internal/testenv"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type nameGetter interface {
	GetNamespace() string
	GetName() string
}

type objectMeta struct {
	GVK  schema.GroupVersionKind
	Name types.NamespacedName
}

type Client struct {
	client.Client

	changed []objectMeta
}

func NewClient(env testenv.Interface, objects ...runtime.Object) *Client {
	c, err := env.NewClient(objects...)

	if err != nil {
		panic(err)
	}

	return &Client{
		Client: c,
	}
}

func (c *Client) addChanged(obj runtime.Object) {
	name := obj.(nameGetter)
	c.changed = append(c.changed, objectMeta{
		GVK: obj.GetObjectKind().GroupVersionKind(),
		Name: types.NamespacedName{
			Namespace: name.GetNamespace(),
			Name:      name.GetName(),
		},
	})
}

func (c *Client) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOptionFunc) error {
	c.addChanged(obj)
	return c.Client.Create(ctx, obj, opts...)
}

func (c *Client) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOptionFunc) error {
	c.addChanged(obj)
	return c.Client.Update(ctx, obj, opts...)
}

func (c *Client) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOptionFunc) error {
	c.addChanged(obj)
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func (c *Client) GetChangedObjects() []runtime.Object {
	// nolint: prealloc
	var objects []runtime.Object
	objectMap := map[string]struct{}{}

	for _, meta := range c.changed {
		key := meta.GVK.String() + " " + meta.Name.String()

		if _, ok := objectMap[key]; ok {
			continue
		}

		obj := new(unstructured.Unstructured)
		obj.SetGroupVersionKind(meta.GVK)

		if err := c.Client.Get(context.Background(), meta.Name, obj); err != nil {
			panic(err)
		}

		objects = append(objects, obj)
		objectMap[key] = struct{}{}
	}

	return objects
}
