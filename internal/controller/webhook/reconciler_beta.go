package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=pullup.dev,resources=resourcetemplates,verbs=get;list;watch;patch
// +kubebuilder:rbac:groups=pullup.dev,resources=httpwebhooks,verbs=get;list;watch
// +kubebuilder:rbac:groups=pullup.dev,resources=githubwebhooks,verbs=get;list;watch

const ownerRefUIDField = "metadata.ownerReferences.uid"

var ErrObjectNotWebhook = errors.New("object is not a webhook")

// BetaReconcilerConfigSet provides a BetaReconcilerConfig.
// nolint: gochecknoglobals
var BetaReconcilerConfigSet = wire.NewSet(
	wire.Struct(new(BetaReconcilerConfig), "*"),
)

// BetaReconcilerFactorySet provides a BetaReconcilerFactory.
// nolint: gochecknoglobals
var BetaReconcilerFactorySet = wire.NewSet(
	BetaReconcilerConfigSet,
	NewBetaReconcilerFactory,
)

type BetaReconcilerConfig struct {
	Client   client.Client
	Logger   Logger
	Recorder record.EventRecorder
}

type BetaReconcilerFactory struct {
	BetaReconcilerConfig
}

func NewBetaReconcilerFactory(conf BetaReconcilerConfig, mgr manager.Manager) (*BetaReconcilerFactory, error) {
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.ResourceTemplate{}, ownerRefUIDField, func(obj runtime.Object) []string {
		var result []string

		for _, ref := range obj.(*v1beta1.ResourceTemplate).OwnerReferences {
			result = append(result, string(ref.UID))
		}

		return result
	})
	if err != nil {
		return nil, fmt.Errorf("index failed: %w", err)
	}

	return &BetaReconcilerFactory{
		BetaReconcilerConfig: conf,
	}, nil
}

func (b *BetaReconcilerFactory) NewReconciler(obj runtime.Object) reconcile.Reconciler {
	return &BetaReconciler{
		BetaReconcilerConfig: b.BetaReconcilerConfig,
		Object:               obj,
	}
}

func (b *BetaReconcilerFactory) Build(mgr manager.Manager, obj runtime.Object) error {
	return builder.
		ControllerManagedBy(mgr).
		For(obj).
		Owns(&v1beta1.ResourceTemplate{}).
		Complete(b.NewReconciler(obj))
}

type BetaReconciler struct {
	BetaReconcilerConfig

	Object runtime.Object
}

func (r *BetaReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	obj := r.Object.DeepCopyObject()
	ctx := context.Background()

	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to get webhook: %w", err)
	}

	logger := r.Logger.WithValues("webhook", obj)
	ctx = logr.NewContext(ctx, logger)

	webhook, ok := obj.(hookutil.Webhook)
	if !ok {
		return reconcile.Result{}, ErrObjectNotWebhook
	}

	list := new(v1beta1.ResourceTemplateList)
	err := r.Client.List(ctx, list, client.InNamespace(webhook.GetNamespace()), client.MatchingFields(map[string]string{
		ownerRefUIDField: string(webhook.GetUID()),
	}))
	if err != nil {
		return reconcile.Result{Requeue: true}, fmt.Errorf("failed to list resource templates: %w", err)
	}

	for _, rt := range list.Items {
		rt := rt
		result := r.patchResource(ctx, webhook, &rt)

		result.RecordEvent(r.Recorder)

		if err := result.Error; err != nil {
			logger.Error(err, result.GetMessage())

			return reconcile.Result{Requeue: result.Requeue}, err
		}

		logger.Info(result.GetMessage())
	}

	return reconcile.Result{}, nil
}

func (r *BetaReconciler) patchResource(ctx context.Context, webhook hookutil.Webhook, rt *v1beta1.ResourceTemplate) controller.Result {
	webhookBuf, err := json.Marshal(webhook)
	if err != nil {
		return controller.Result{
			Object: webhook,
			Error:  fmt.Errorf("failed to marshal webhook: %w", err),
			Reason: ReasonPatchFailed,
		}
	}

	patchesBuf, err := json.Marshal(webhook.GetSpec().Patches)
	if err != nil {
		return controller.Result{
			Object: webhook,
			Error:  fmt.Errorf("failed to marshal patches: %w", err),
			Reason: ReasonPatchFailed,
		}
	}

	patch, err := json.Marshal([]v1beta1.JSONPatch{
		{
			Operation: "replace",
			Path:      "/spec/patches",
			Value:     &extv1.JSON{Raw: patchesBuf},
		},
		{
			Operation: "replace",
			Path:      "/spec/data/webhook",
			Value:     &extv1.JSON{Raw: webhookBuf},
		},
	})
	if err != nil {
		return controller.Result{
			Object: webhook,
			Error:  fmt.Errorf("failed to marshal json patch: %w", err),
			Reason: ReasonPatchFailed,
		}
	}

	if err := r.Client.Patch(ctx, rt, client.RawPatch(types.JSONPatchType, patch)); err != nil {
		return controller.Result{
			Object:  webhook,
			Error:   fmt.Errorf("failed to patch resource template: %w", err),
			Reason:  ReasonPatchFailed,
			Requeue: true,
		}
	}

	return controller.Result{
		Object:  webhook,
		Message: fmt.Sprintf("Patched resource template %q", rt.Name),
		Reason:  ReasonPatched,
	}
}
