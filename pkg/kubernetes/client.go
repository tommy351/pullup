package kubernetes

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/mitchellh/go-homedir"
	"github.com/tommy351/pullup/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type configGetter func() (*rest.Config, error)

type Client struct {
	config    *rest.Config
	namespace string
}

func NewClient(conf *config.KubernetesConfig) (*Client, error) {
	var (
		restConf *rest.Config
		err      error
	)

	getters := []configGetter{
		rest.InClusterConfig,
		readConfigFromHomeDir,
	}

	for _, getter := range getters {
		restConf, err = getter()

		if err == nil {
			break
		}
	}

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

func readConfigFromPath(path string) configGetter {
	return func() (*rest.Config, error) {
		return clientcmd.BuildConfigFromFlags("", path)
	}
}

func readConfigFromHomeDir() (*rest.Config, error) {
	dir, err := homedir.Dir()

	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, ".kube", "config")
	return readConfigFromPath(path)()
}

func (c *Client) newRESTClient(apiVersion string) (rest.Interface, error) {
	var gv schema.GroupVersion
	config := *c.config
	parts := strings.SplitN(apiVersion, "/", 2)

	if len(parts) == 2 {
		gv.Group = parts[0]
		gv.Version = parts[1]
	} else {
		gv.Version = parts[0]
	}

	config.GroupVersion = &gv

	if gv.Group == "" {
		config.APIPath = "/api"
	} else {
		config.APIPath = "/apis"
	}

	return rest.RESTClientFor(&config)
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
