package kubernetes

import (
	"context"
	"strings"

	"github.com/ansel1/merry"
	"github.com/jinzhu/inflection"
	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/cache"
	"github.com/tommy351/pullup/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	// Load auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// nolint: gochecknoglobals
var pluralKindCache = cache.NewMap()

func getPluralKind(kind string) string {
	v, _ := pluralKindCache.LoadOrStore(kind, func() interface{} {
		return inflection.Plural(kind)
	})

	return v.(string)
}

type Client interface {
	Apply(ctx context.Context, resource *Resource) error
	Delete(ctx context.Context, resource *Resource) error
}

type client struct {
	config    *rest.Config
	namespace string
	client    dynamic.Interface
}

func NewClient(conf *config.KubernetesConfig) (Client, error) {
	restConf, err := LoadConfig()

	if err != nil {
		return nil, merry.Wrap(err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConf)

	if err != nil {
		return nil, merry.Wrap(err)
	}

	return &client{
		config:    restConf,
		namespace: conf.Namespace,
		client:    dynamicClient,
	}, nil
}

func (c *client) newResource(resource *Resource) dynamic.ResourceInterface {
	gvr := schema.GroupVersionResource{
		Resource: strings.ToLower(getPluralKind(resource.Kind)),
	}

	parts := strings.SplitN(resource.APIVersion, "/", 2)

	if len(parts) == 2 {
		gvr.Group = parts[0]
		gvr.Version = parts[1]
	} else {
		gvr.Version = parts[0]
	}

	return c.client.Resource(gvr).Namespace(c.namespace)
}

func (c *client) Apply(ctx context.Context, resource *Resource) error {
	logger := zerolog.Ctx(ctx)
	result, err := c.newResource(resource).Get(resource.Name, metav1.GetOptions{})

	if err != nil {
		return merry.Wrap(err)
	}

	resource.OriginalResource = result

	logger.Debug().
		Interface("resource", result.Object).
		Msg("Original resource get")

	if err := c.create(ctx, resource); err == nil {
		return nil
	} else if !IsAlreadyExistError(err) {
		return merry.Wrap(err)
	}

	return c.update(ctx, resource)
}

func (c *client) update(ctx context.Context, resource *Resource) error {
	logger := zerolog.Ctx(ctx)
	name := resource.ModifiedName()
	applied, err := c.newResource(resource).Get(name, metav1.GetOptions{})

	if err != nil {
		return merry.Wrap(err)
	}

	resource.AppliedResource = applied

	if err := resource.Patch(); err != nil {
		return merry.Wrap(err)
	}

	logger.Debug().
		Interface("resource", resource.PatchedResource.Object).
		Msg("Resource to update")

	result, err := c.newResource(resource).
		Update(resource.PatchedResource, metav1.UpdateOptions{})

	if err != nil {
		return merry.Wrap(err)
	}

	logger.Debug().
		Interface("result", result.Object).
		Msg("Patch resource updated")

	return nil
}

func (c *client) create(ctx context.Context, resource *Resource) error {
	logger := zerolog.Ctx(ctx)

	if err := resource.Patch(); err != nil {
		return merry.Wrap(err)
	}

	logger.Debug().
		Interface("resource", resource.PatchedResource.Object).
		Msg("Resource to create")

	result, err := c.newResource(resource).
		Create(resource.PatchedResource, metav1.CreateOptions{})

	if err != nil {
		return merry.Wrap(err)
	}

	logger.Debug().
		Interface("result", result.Object).
		Msg("Patched resource created")

	return nil
}

func (c *client) Delete(ctx context.Context, resource *Resource) error {
	logger := zerolog.Ctx(ctx)
	name := resource.ModifiedName()

	err := c.newResource(resource).
		Delete(resource.ModifiedName(), &metav1.DeleteOptions{})

	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}

		return merry.Wrap(err)
	}

	logger.Debug().
		Str("name", name).
		Str("apiVersion", resource.APIVersion).
		Str("kind", resource.Kind).
		Msg("Patched resource deleted")

	return nil
}
