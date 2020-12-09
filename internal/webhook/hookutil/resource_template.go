package hookutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/jsonutil"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/template"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=pullup.dev,resources=resourcetemplates,verbs=create;patch;delete

const (
	ReasonCreated        = "Created"
	ReasonCreateFailed   = "CreateFailed"
	ReasonUpdated        = "Updated"
	ReasonUpdateFailed   = "UpdateFailed"
	ReasonDeleted        = "Deleted"
	ReasonDeleteFailed   = "DeleteFailed"
	ReasonInvalidWebhook = "InvalidWebhook"
	ReasonNotExist       = "NotExist"
)

type ResourceTemplateAction string

const (
	ActionApply  ResourceTemplateAction = "apply"
	ActionDelete ResourceTemplateAction = "delete"
)

var (
	ErrInvalidResourceTemplateAction = errors.New("invalid action")
	ErrResourceNameRequired          = errors.New("resourceName is required")
)

type ResourceTemplateOptions struct {
	Action              ResourceTemplateAction
	Event               interface{}
	Webhook             Webhook
	DefaultResourceName string
}

// ResourceTemplateClientSet provides the client for ResourceTemplate.
// nolint: gochecknoglobals
var ResourceTemplateClientSet = wire.NewSet(
	wire.Struct(new(ResourceTemplateClient), "*"),
)

type Webhook interface {
	runtime.Object

	GetName() string
	GetNamespace() string
	GetUID() types.UID
	GetSpec() v1beta1.WebhookSpec
}

type ResourceTemplateClient struct {
	Client   client.Client
	Recorder record.EventRecorder
}

func (r *ResourceTemplateClient) Handle(ctx context.Context, options *ResourceTemplateOptions) error {
	var result controller.Result
	logger := logr.FromContextOrDiscard(ctx)
	rt, err := r.generateResourceTemplate(options)
	if err != nil {
		return err
	}

	switch options.Action {
	case ActionApply:
		result = r.apply(ctx, rt, options)
	case ActionDelete:
		result = r.delete(ctx, rt, options)
	default:
		return ErrInvalidResourceTemplateAction
	}

	result.RecordEvent(r.Recorder)

	if err := result.Error; err != nil {
		logger.Error(err, result.GetMessage())

		return err
	}

	logger.Info(result.GetMessage())

	return nil
}

func (r *ResourceTemplateClient) generateResourceTemplate(options *ResourceTemplateOptions) (*v1beta1.ResourceTemplate, error) {
	apiVersion, kind := options.Webhook.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	rt := &v1beta1.ResourceTemplate{
		TypeMeta: k8s.GVKToTypeMeta(v1beta1.GroupVersion.WithKind("ResourceTemplate")),
		ObjectMeta: metav1.ObjectMeta{
			Namespace: options.Webhook.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         apiVersion,
					Kind:               kind,
					Name:               options.Webhook.GetName(),
					UID:                options.Webhook.GetUID(),
					Controller:         pointer.BoolPtr(true),
					BlockOwnerDeletion: pointer.BoolPtr(true),
				},
			},
		},
		Spec: v1beta1.ResourceTemplateSpec{
			WebhookRef: &v1beta1.ObjectReference{
				APIVersion: apiVersion,
				Kind:       kind,
				Name:       options.Webhook.GetName(),
			},
			Patches: options.Webhook.GetSpec().Patches,
		},
	}

	buf, err := json.Marshal(map[string]interface{}{
		v1beta1.DataKeyEvent: options.Event,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource template data: %w", err)
	}

	rt.Spec.Data = extv1.JSON{Raw: buf}
	name, err := r.renderResourceName(rt, options)
	if err != nil {
		return nil, err
	}

	rt.Name = name

	return rt, nil
}

func (r *ResourceTemplateClient) renderResourceName(rt *v1beta1.ResourceTemplate, options *ResourceTemplateOptions) (string, error) {
	resourceName := options.Webhook.GetSpec().ResourceName
	if resourceName == "" {
		resourceName = options.DefaultResourceName
	}

	if resourceName == "" {
		return "", ErrResourceNameRequired
	}

	buf, err := jsonutil.AddMapKey(rt.Spec.Data.Raw, v1beta1.DataKeyWebhook, options.Webhook)
	if err != nil {
		return "", fmt.Errorf("failed to mutate json: %w", err)
	}

	name, err := template.RenderFromJSON(resourceName, extv1.JSON{Raw: buf})
	if err != nil {
		return "", fmt.Errorf("failed to generate name: %w", err)
	}

	return name, nil
}

func (r *ResourceTemplateClient) apply(ctx context.Context, rt *v1beta1.ResourceTemplate, options *ResourceTemplateOptions) controller.Result {
	if err := r.Client.Create(ctx, rt); err == nil {
		return controller.Result{
			Object:  options.Webhook,
			Message: fmt.Sprintf("Created resource template: %s", rt.Name),
			Reason:  ReasonCreated,
		}
	} else if !kerrors.IsAlreadyExists(err) {
		return controller.Result{
			Object: options.Webhook,
			Error:  fmt.Errorf("failed to create resource template: %w", err),
			Reason: ReasonCreateFailed,
		}
	}

	patchValue, err := json.Marshal(rt.Spec)
	if err != nil {
		return controller.Result{
			Object: options.Webhook,
			Error:  fmt.Errorf("failed to marshal patch value: %w", err),
			Reason: ReasonUpdateFailed,
		}
	}

	patch, err := json.Marshal([]v1beta1.JSONPatch{
		{
			Operation: "replace",
			Path:      "/spec",
			Value:     &extv1.JSON{Raw: patchValue},
		},
	})
	if err != nil {
		return controller.Result{
			Object: options.Webhook,
			Error:  fmt.Errorf("failed to marshal resource template spec: %w", err),
			Reason: ReasonUpdateFailed,
		}
	}

	if err := r.Client.Patch(ctx, rt, client.RawPatch(types.JSONPatchType, patch)); err != nil {
		return controller.Result{
			Object: options.Webhook,
			Error:  fmt.Errorf("failed to patch resource template: %w", err),
			Reason: ReasonUpdateFailed,
		}
	}

	return controller.Result{
		Object:  options.Webhook,
		Message: fmt.Sprintf("Updated resource template: %s", rt.Name),
		Reason:  ReasonUpdated,
	}
}

func (r *ResourceTemplateClient) delete(ctx context.Context, rt *v1beta1.ResourceTemplate, options *ResourceTemplateOptions) controller.Result {
	if err := r.Client.Delete(ctx, rt); err != nil {
		if kerrors.IsNotFound(err) {
			return controller.Result{
				Object:  options.Webhook,
				Message: fmt.Sprintf("Resource template does not exist: %s", rt.Name),
				Reason:  ReasonNotExist,
			}
		}

		return controller.Result{
			Object: options.Webhook,
			Error:  fmt.Errorf("failed to delete resource template: %w", err),
			Reason: ReasonDeleteFailed,
		}
	}

	return controller.Result{
		Object:  options.Webhook,
		Message: fmt.Sprintf("Deleted resource template: %s", rt.Name),
		Reason:  ReasonDeleted,
	}
}
