package kubernetes

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ansel1/merry"
	"github.com/jinzhu/inflection"
	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/cache"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func normalizeJSONValue(input interface{}) (interface{}, error) {
	v := reflect.ValueOf(input)

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), nil

	case reflect.Float32, reflect.Float64:
		return v.Float(), nil

	case reflect.Slice, reflect.Array:
		arr := make([]interface{}, v.Len())

		for i := 0; i < v.Len(); i++ {
			value, err := normalizeJSONValue(v.Index(i).Interface())

			if err != nil {
				return nil, merry.Wrap(err)
			}

			arr[i] = value
		}

		return arr, nil

	case reflect.Map:
		result := map[string]interface{}{}

		for _, key := range v.MapKeys() {
			value, err := normalizeJSONValue(v.MapIndex(key).Interface())

			if err != nil {
				return nil, merry.Wrap(err)
			}

			result[key.Interface().(string)] = value
		}

		return result, nil
	}

	return nil, merry.Wrap(fmt.Errorf("unsupported type: %T", input))
}

type Client interface {
	Apply(ctx context.Context, resource *model.Resource) error
	Delete(ctx context.Context, resource *model.Resource) error
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

func (c *client) newResource(resource *model.Resource) dynamic.ResourceInterface {
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

	ns := resource.Namespace

	if ns != "" {
		ns = c.namespace
	}

	return c.client.Resource(gvr).Namespace(ns)
}

func (c *client) Apply(ctx context.Context, resource *model.Resource) error {
	logger := zerolog.Ctx(ctx)

	if create := resource.Create; create != nil {
		obj, err := normalizeJSONValue(create)

		if err != nil {
			return merry.Wrap(err)
		}

		resource.OriginalResource = &unstructured.Unstructured{
			Object: obj.(map[string]interface{}),
		}

		resource.OriginalResource.SetAPIVersion(resource.APIVersion)
		resource.OriginalResource.SetKind(resource.Kind)
		resource.OriginalResource.SetName(resource.Name)
		resource.OriginalResource.SetCreationTimestamp(metav1.Time{Time: time.Now()})
		resource.OriginalResource.SetResourceVersion("0")
		resource.OriginalResource.SetSelfLink("0")
		resource.OriginalResource.SetUID("0")
	} else {
		result, err := c.newResource(resource).Get(resource.Name, metav1.GetOptions{})

		if err != nil {
			return merry.Wrap(err)
		}

		resource.OriginalResource = result

		logger.Debug().
			Interface("resource", result.Object).
			Msg("Original resource get")
	}

	if err := c.create(ctx, resource); err == nil {
		return nil
	} else if !IsAlreadyExistError(err) {
		return merry.Wrap(err)
	}

	return c.update(ctx, resource)
}

func (c *client) update(ctx context.Context, resource *model.Resource) error {
	logger := zerolog.Ctx(ctx)
	name := resource.ModifiedName()
	applied, err := c.newResource(resource).Get(name, metav1.GetOptions{})

	if err != nil {
		return merry.Wrap(err)
	}

	resource.AppliedResource = applied

	if err := Patch(resource); err != nil {
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

func (c *client) create(ctx context.Context, resource *model.Resource) error {
	logger := zerolog.Ctx(ctx)

	if err := Patch(resource); err != nil {
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

func (c *client) Delete(ctx context.Context, resource *model.Resource) error {
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
