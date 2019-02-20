package kubernetes

import (
	"context"

	"github.com/jinzhu/inflection"
	"github.com/tommy351/pullup/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
)

type Client struct {
	config    *rest.Config
	namespace string
}

func NewClient(conf *config.KubernetesConfig) (*Client, error) {
	restConf, err := LoadConfig()

	if err != nil {
		return nil, err
	}

	restConf.UserAgent = rest.DefaultKubernetesUserAgent()
	restConf.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	return &Client{
		config:    restConf,
		namespace: conf.Namespace,
	}, nil
}

func (c *Client) newRESTClient(apiVersion string) (rest.Interface, error) {
	return rest.RESTClientFor(GetVersionedConfig(c.config, apiVersion))
}

func (c *Client) newRequest(ctx context.Context, client rest.Interface, verb string) *rest.Request {
	return client.Verb(verb).Context(ctx).Namespace(c.namespace)
}

func (c *Client) Apply(ctx context.Context, resource *Resource) error {
	client, err := c.newRESTClient(resource.APIVersion)

	if err != nil {
		return err
	}

	kind := inflection.Plural(resource.Kind)

	// Get the original resource
	result := c.newRequest(ctx, client, "GET").
		Resource(kind).
		Name(resource.Name).
		Do()

	if err := result.Error(); err != nil {
		return err
	}

	// Modify the resource
	modified, err := PatchResource(result, resource)

	if err != nil {
		return err
	}

	// Apply the modified resource
	result = c.newRequest(ctx, client, "PUT").
		Resource(kind).
		Name(resource.ModifiedName()).
		Body(modified).
		Do()

	if err := result.Error(); err == nil {
		return nil
	} else if !errors.IsNotFound(err) {
		return err
	}

	// Create the resource if apply failed
	result = c.newRequest(ctx, client, "POST").
		Resource(kind).
		Body(modified).
		Do()

	return result.Error()
}

func (c *Client) Delete(ctx context.Context, resource *Resource) error {
	client, err := c.newRESTClient(resource.APIVersion)

	if err != nil {
		return err
	}

	kind := inflection.Plural(resource.Kind)
	result := c.newRequest(ctx, client, "DELETE").
		Resource(kind).
		Name(resource.ModifiedName()).
		Do()

	if err := result.Error(); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	return nil
}
