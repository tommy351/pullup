package kubernetes

import (
	"context"

	"github.com/ansel1/merry"
	"github.com/jinzhu/inflection"
	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/cache"
	"github.com/tommy351/pullup/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
)

var pluralKindCache = cache.NewMap()

func getPluralKind(kind string) string {
	v, _ := pluralKindCache.LoadOrStore(kind, func() interface{} {
		return inflection.Plural(kind)
	})

	return v.(string)
}

type Client struct {
	config      *rest.Config
	namespace   string
	clientCache *cache.Map
}

func NewClient(conf *config.KubernetesConfig) (*Client, error) {
	restConf, err := LoadConfig()

	if err != nil {
		return nil, merry.Wrap(err)
	}

	restConf.UserAgent = rest.DefaultKubernetesUserAgent()
	restConf.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	return &Client{
		config:      restConf,
		namespace:   conf.Namespace,
		clientCache: cache.NewMap(),
	}, nil
}

func (c *Client) newRESTClient(apiVersion string) (rest.Interface, error) {
	var err error

	client, _ := c.clientCache.LoadOrStore(apiVersion, func() interface{} {
		var value rest.Interface
		value, err = rest.RESTClientFor(GetVersionedConfig(c.config, apiVersion))
		return value
	})

	if err != nil {
		return nil, merry.Wrap(err)
	}

	return client.(rest.Interface), nil
}

func (c *Client) newRequest(ctx context.Context, client rest.Interface, verb string) *rest.Request {
	return client.Verb(verb).Context(ctx).Namespace(c.namespace)
}

func (c *Client) logResult(ctx context.Context, result rest.Result) *zerolog.Event {
	logger := zerolog.Ctx(ctx)
	event := logger.Debug()

	if event.Enabled() {
		raw, _ := result.Raw()
		event = event.RawJSON("result", raw)
	}

	return event
}

func (c *Client) Apply(ctx context.Context, resource *Resource) error {
	logger := zerolog.Ctx(ctx)
	client, err := c.newRESTClient(resource.APIVersion)

	if err != nil {
		return merry.Wrap(err)
	}

	kind := getPluralKind(resource.Kind)

	// Get the original resource
	result := c.newRequest(ctx, client, "GET").
		Resource(kind).
		Name(resource.Name).
		Do()

	if err := result.Error(); err != nil {
		return merry.Wrap(err)
	}

	c.logResult(ctx, result).Msg("Original resource get")

	// Patch the resource
	patched, err := PatchResource(result, resource)

	if err != nil {
		return merry.Wrap(err)
	}

	logger.Debug().RawJSON("resource", patched).Msg("Original resource patched")

	// Apply the patched resource
	result = c.newRequest(ctx, client, "PUT").
		Resource(kind).
		Name(resource.ModifiedName()).
		Body(patched).
		Do()

	if err := result.Error(); err == nil {
		c.logResult(ctx, result).Msg("Patched resource applied")
		return nil
	} else if !errors.IsNotFound(err) {
		return merry.Wrap(err)
	}

	// Create the resource if apply failed
	result = c.newRequest(ctx, client, "POST").
		Resource(kind).
		Body(patched).
		Do()

	if err := result.Error(); err != nil {
		return merry.Wrap(err)
	}

	c.logResult(ctx, result).Msg("Patched resource created")
	return nil
}

func (c *Client) Delete(ctx context.Context, resource *Resource) error {
	client, err := c.newRESTClient(resource.APIVersion)

	if err != nil {
		return merry.Wrap(err)
	}

	result := c.newRequest(ctx, client, "DELETE").
		Resource(getPluralKind(resource.Kind)).
		Name(resource.ModifiedName()).
		Do()

	if err := result.Error(); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return merry.Wrap(err)
	}

	c.logResult(ctx, result).Msg("Patched resource deleted")
	return nil
}
