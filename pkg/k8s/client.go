package k8s

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/client/clientset/versioned"
	"github.com/tommy351/pullup/pkg/client/informers/externalversions"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	// Load auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type jsonPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type Client struct {
	Namespace string
	Dynamic   dynamic.Interface
	Client    versioned.Interface
}

func NewClient(config *Config) (*Client, error) {
	conf, err := LoadConfig(config)

	if err != nil {
		return nil, xerrors.Errorf("failed to load kubernetes config: %w", err)
	}

	c := &Client{
		Namespace: config.Namespace,
	}

	if c.Dynamic, err = dynamic.NewForConfig(conf); err != nil {
		return nil, xerrors.Errorf("failed to create a dynamic client: %w", err)
	}

	if c.Client, err = versioned.NewForConfig(conf); err != nil {
		return nil, xerrors.Errorf("failed to create a versioned client: %w", err)
	}

	return c, nil
}

func (c *Client) GetWebhook(_ context.Context, name string) (*v1alpha1.Webhook, error) {
	webhook, err := c.Client.PullupV1alpha1().Webhooks(c.Namespace).Get(name, metav1.GetOptions{})

	if err != nil {
		return nil, xerrors.Errorf("failed to get webhook: %w", err)
	}

	webhook.SetGroupVersionKind(v1alpha1.Kind("Webhook"))

	return webhook, nil
}

func (c *Client) ApplyResourceSet(ctx context.Context, rs *v1alpha1.ResourceSet) error {
	logger := zerolog.Ctx(ctx).With().
		Str("namespace", rs.Namespace).
		Str("name", rs.Name).
		Logger()

	client := c.Client.PullupV1alpha1().ResourceSets(c.Namespace)

	if _, err := client.Create(rs); err == nil {
		logger.Debug().Msg("Created resource set")
		return nil
	} else if !IsAlreadyExistError(err) {
		return xerrors.Errorf("failed to create resource set: %w", err)
	}

	patchData, err := json.Marshal([]jsonPatchOp{
		{
			Op:    "replace",
			Path:  "/spec",
			Value: rs.Spec,
		},
	})

	if err != nil {
		return xerrors.Errorf("failed to marshal resource set spec: %w", err)
	}

	if _, err := client.Patch(rs.Name, types.JSONPatchType, patchData); err != nil {
		return xerrors.Errorf("failed to patch resource set: %w", err)
	}

	logger.Debug().Msg("Updated resource set")
	return nil
}

func (c *Client) DeleteResourceSet(ctx context.Context, name string) error {
	logger := zerolog.Ctx(ctx).With().
		Str("namespace", c.Namespace).
		Str("name", name).
		Logger()

	if err := c.Client.PullupV1alpha1().ResourceSets(c.Namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return xerrors.Errorf("failed to delete resource set: %w", err)
	}

	logger.Debug().Msg("Deleted resource set")
	return nil
}

func (c *Client) NewInformer(ctx context.Context) externalversions.SharedInformerFactory {
	return externalversions.NewSharedInformerFactoryWithOptions(
		c.Client,
		0,
		externalversions.WithNamespace(c.Namespace),
	)
}

func (c *Client) NewDynamicInterface(ctx context.Context, gvr schema.GroupVersionResource) dynamic.ResourceInterface {
	return c.Dynamic.Resource(gvr).Namespace(c.Namespace)
}
