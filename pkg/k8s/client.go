package k8s

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock

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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Load auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Config struct {
	Namespace string `mapstructure:"namespace"`
	Config    string `mapstructure:"config"`
}

type Client interface {
	GetWebhook(ctx context.Context, name string) (*v1alpha1.Webhook, error)
	ApplyResourceSet(ctx context.Context, rs *v1alpha1.ResourceSet) error
	DeleteResourceSet(ctx context.Context, name string) error
	NewInformer(ctx context.Context) externalversions.SharedInformerFactory
	NewDynamicInterface(ctx context.Context, gvr schema.GroupVersionResource) dynamic.ResourceInterface
}

type client struct {
	namespace string
	dynamic   dynamic.Interface
	client    versioned.Interface
}

func NewClient(config *Config) (Client, error) {
	var (
		conf *rest.Config
		err  error
	)

	if path := config.Config; path == "" {
		conf, err = rest.InClusterConfig()
	} else {
		conf, err = clientcmd.BuildConfigFromFlags("", path)
	}

	if err != nil {
		return nil, xerrors.Errorf("failed to load kubernetes config: %w", err)
	}

	c := &client{
		namespace: config.Namespace,
	}

	if c.dynamic, err = dynamic.NewForConfig(conf); err != nil {
		return nil, xerrors.Errorf("failed to create a dynamic client: %w", err)
	}

	if c.client, err = versioned.NewForConfig(conf); err != nil {
		return nil, xerrors.Errorf("failed to create a versioned client: %w", err)
	}

	return c, nil
}

func (c *client) GetWebhook(_ context.Context, name string) (*v1alpha1.Webhook, error) {
	webhook, err := c.client.PullupV1alpha1().Webhooks(c.namespace).Get(name, metav1.GetOptions{})

	if err != nil {
		return nil, xerrors.Errorf("failed to get webhook: %w", err)
	}

	webhook.APIVersion = v1alpha1.SchemeGroupVersion.String()
	webhook.Kind = "Webhook"
	return webhook, nil
}

func (c *client) ApplyResourceSet(ctx context.Context, rs *v1alpha1.ResourceSet) error {
	logger := zerolog.Ctx(ctx).With().
		Str("namespace", rs.Namespace).
		Str("name", rs.Name).
		Logger()

	client := c.client.PullupV1alpha1().ResourceSets(c.namespace)

	if _, err := client.Create(rs); err == nil {
		logger.Debug().Msg("Created resource set")
		return nil
	} else if !IsAlreadyExistError(err) {
		return xerrors.Errorf("failed to create resource set: %w", err)
	}

	patchData, err := json.Marshal(map[string]interface{}{
		"spec": rs.Spec,
	})

	if err != nil {
		return xerrors.Errorf("failed to marshal resource set spec: %w", err)
	}

	if _, err := client.Patch(rs.Name, types.MergePatchType, patchData); err != nil {
		return xerrors.Errorf("failed to patch resource set: %w", err)
	}

	logger.Debug().Msg("Updated resource set")
	return nil
}

func (c *client) DeleteResourceSet(ctx context.Context, name string) error {
	logger := zerolog.Ctx(ctx).With().
		Str("namespace", c.namespace).
		Str("name", name).
		Logger()

	if err := c.client.PullupV1alpha1().ResourceSets(c.namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return xerrors.Errorf("failed to delete resource set: %w", err)
	}

	logger.Debug().Msg("Deleted resource set")
	return nil
}

func (c *client) NewInformer(ctx context.Context) externalversions.SharedInformerFactory {
	return externalversions.NewSharedInformerFactory(c.client, 0)
}

func (c *client) NewDynamicInterface(ctx context.Context, gvr schema.GroupVersionResource) dynamic.ResourceInterface {
	return c.dynamic.Resource(gvr).Namespace(c.namespace)
}
