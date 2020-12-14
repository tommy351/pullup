package hookutil

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/jsonutil"
	"github.com/tommy351/pullup/internal/template"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=pullup.dev,resources=triggers,verbs=get;list;watch
// +kubebuilder:rbac:groups=pullup.dev,resources=resourcetemplates,verbs=create;patch;delete

const (
	ReasonCreated        = "Created"
	ReasonCreateFailed   = "CreateFailed"
	ReasonAlreadyExists  = "AlreadyExists"
	ReasonUpdated        = "Updated"
	ReasonUpdateFailed   = "UpdateFailed"
	ReasonDeleted        = "Deleted"
	ReasonDeleteFailed   = "DeleteFailed"
	ReasonInvalidWebhook = "InvalidWebhook"
	ReasonNotExist       = "NotExist"
	ReasonTriggered      = "Triggered"
	ReasonTriggerFailed  = "TriggerFailed"
)

// TriggerHandlerSet provides a TriggerHandler.
// nolint: gochecknoglobals
var TriggerHandlerSet = wire.NewSet(
	wire.Struct(new(TriggerHandler), "*"),
)

type TriggerSource interface {
	runtime.Object
	metav1.Object
}

type RenderedTrigger struct {
	ResourceTemplate *v1beta1.ResourceTemplate
	Trigger          *v1beta1.Trigger
}

type TriggerOptions struct {
	Source        TriggerSource
	Triggers      []v1beta1.EventSourceTrigger
	DefaultAction v1beta1.Action
	Action        v1beta1.Action
	Event         interface{}
}

type TriggerHandler struct {
	Client   client.Client
	Recorder record.EventRecorder
}

func (t *TriggerHandler) Handle(ctx context.Context, options *TriggerOptions) error {
	action, err := t.renderAction(options)
	if err != nil {
		return err
	}

	triggers := make([]*RenderedTrigger, len(options.Triggers))

	for i, trigger := range options.Triggers {
		trigger := trigger
		rt, err := t.renderTrigger(ctx, &trigger, options)
		if err != nil {
			return err
		}

		triggers[i] = rt
	}

	for _, trigger := range triggers {
		trigger := trigger

		if err := t.handleTrigger(ctx, trigger, action, options); err != nil {
			return err
		}
	}

	return nil
}

func (t *TriggerHandler) renderAction(options *TriggerOptions) (v1beta1.Action, error) {
	action := options.Action
	if action == "" {
		action = options.DefaultAction
	}

	buf, err := json.Marshal(map[string]interface{}{
		v1beta1.DataKeyEvent:  options.Event,
		v1beta1.DataKeyAction: options.DefaultAction,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	result, err := template.RenderFromJSON(string(action), extv1.JSON{Raw: buf})
	if err != nil {
		return "", fmt.Errorf("failed to render action: %w", err)
	}

	action = v1beta1.Action(result)

	if !v1beta1.IsActionValid(action) {
		return "", ErrInvalidAction
	}

	return action, nil
}

func (t *TriggerHandler) renderTrigger(ctx context.Context, st *v1beta1.EventSourceTrigger, options *TriggerOptions) (*RenderedTrigger, error) {
	trigger := new(v1beta1.Trigger)
	triggerKey := st.Ref.NamespacedName()

	if triggerKey.Namespace == "" {
		triggerKey.Namespace = options.Source.GetNamespace()
	}

	if err := t.Client.Get(ctx, triggerKey, trigger); err != nil {
		if kerrors.IsNotFound(err) {
			return nil, TriggerNotFoundError{key: triggerKey, err: err}
		}

		return nil, fmt.Errorf("failed to get Trigger: %w", err)
	}

	result := &RenderedTrigger{
		Trigger: trigger,
		ResourceTemplate: &v1beta1.ResourceTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: trigger.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         trigger.APIVersion,
						Kind:               trigger.Kind,
						Name:               trigger.Name,
						UID:                trigger.UID,
						Controller:         pointer.BoolPtr(true),
						BlockOwnerDeletion: pointer.BoolPtr(true),
					},
				},
			},
			Spec: v1beta1.ResourceTemplateSpec{
				TriggerRef: &v1beta1.ObjectReference{
					APIVersion: trigger.APIVersion,
					Kind:       trigger.Kind,
					Namespace:  trigger.Namespace,
					Name:       trigger.Name,
				},
				Patches: trigger.Spec.Patches,
			},
		},
	}

	data, err := t.renderData(st, trigger, options)
	if err != nil {
		return nil, err
	}

	if result.ResourceTemplate.Name, err = t.renderName(trigger, data); err != nil {
		return nil, err
	}

	if result.ResourceTemplate.Spec.Data, err = t.finalizeData(data); err != nil {
		return nil, err
	}

	return result, nil
}

func (t *TriggerHandler) renderData(st *v1beta1.EventSourceTrigger, trigger *v1beta1.Trigger, options *TriggerOptions) (extv1.JSON, error) {
	eventBuf, err := json.Marshal(options.Event)
	if err != nil {
		return extv1.JSON{}, fmt.Errorf("failed to marshal event data: %w", err)
	}

	render := func() (extv1.JSON, error) {
		buf, err := json.Marshal(map[string]interface{}{
			v1beta1.DataKeyEvent:   extv1.JSON{Raw: eventBuf},
			v1beta1.DataKeyTrigger: trigger,
		})
		if err != nil {
			return extv1.JSON{}, fmt.Errorf("failed to marshal data: %w", err)
		}

		return extv1.JSON{Raw: buf}, nil
	}

	if st.Transform != nil && st.Transform.Raw != nil {
		buf, err := render()
		if err != nil {
			return extv1.JSON{}, nil
		}

		transformed, err := template.RenderFromJSON(string(st.Transform.Raw), buf)
		if err != nil {
			return extv1.JSON{}, fmt.Errorf("failed to render transform data: %w", err)
		}

		eventBuf = []byte(transformed)
	}

	result, err := ValidateJSONSchema(trigger.Spec.Schema, &extv1.JSON{Raw: eventBuf})
	if err != nil {
		return extv1.JSON{}, err
	}

	eventBuf = result.Raw

	return render()
}

func (t *TriggerHandler) renderName(trigger *v1beta1.Trigger, data extv1.JSON) (string, error) {
	name, err := template.RenderFromJSON(trigger.Spec.ResourceName, data)
	if err != nil {
		return "", fmt.Errorf("failed to render name: %w", err)
	}

	if e := validation.NameIsDNSSubdomain(name, false); len(e) > 0 {
		return "", ValidationErrors(e)
	}

	return name, nil
}

func (t *TriggerHandler) finalizeData(data extv1.JSON) (extv1.JSON, error) {
	buf, err := jsonutil.PickKeys(data.Raw, []string{v1beta1.DataKeyEvent})
	if err != nil {
		return extv1.JSON{}, fmt.Errorf("failed to pick json keys: %w", err)
	}

	return extv1.JSON{Raw: buf}, nil
}

func (t *TriggerHandler) handleTrigger(ctx context.Context, trigger *RenderedTrigger, action v1beta1.Action, options *TriggerOptions) error {
	var result controller.Result
	logger := logr.FromContextOrDiscard(ctx)

	switch action {
	case v1beta1.ActionCreate:
		result = t.createResource(ctx, trigger.ResourceTemplate)
	case v1beta1.ActionUpdate:
		result = t.updateResource(ctx, trigger.ResourceTemplate)
	case v1beta1.ActionApply:
		result = t.applyResource(ctx, trigger.ResourceTemplate)
	case v1beta1.ActionDelete:
		result = t.deleteResource(ctx, trigger.ResourceTemplate)
	default:
		return ErrInvalidAction
	}

	result.RecordEvent(t.Recorder, trigger.Trigger)

	t.recordSourceEvent(options.Source, &result, trigger.ResourceTemplate.Spec.TriggerRef)

	if err := result.Error; err != nil {
		logger.Error(err, result.GetMessage())

		return err
	}

	logger.Info(result.GetMessage())

	return nil
}

func (t *TriggerHandler) recordSourceEvent(object runtime.Object, input *controller.Result, triggerRef *v1beta1.ObjectReference) {
	var r controller.Result

	if err := input.Error; err != nil {
		r = controller.Result{
			Error:  fmt.Errorf("trigger failed: %w", err),
			Reason: ReasonTriggerFailed,
		}
	} else {
		r = controller.Result{
			Message: fmt.Sprintf("Triggered: %s", triggerRef.NamespacedName()),
			Reason:  ReasonTriggered,
		}
	}

	r.RecordEvent(t.Recorder, object)
}

func (t *TriggerHandler) createResource(ctx context.Context, rt *v1beta1.ResourceTemplate) controller.Result {
	if err := t.Client.Create(ctx, rt); err != nil {
		if kerrors.IsAlreadyExists(err) {
			return controller.Result{
				Message: fmt.Sprintf("Resource template already exists: %s", rt.Name),
				Reason:  ReasonAlreadyExists,
			}
		}

		return controller.Result{
			Error:  fmt.Errorf("failed to create resource template: %w", err),
			Reason: ReasonCreateFailed,
		}
	}

	return controller.Result{
		Message: fmt.Sprintf("Created resource template: %s", rt.Name),
		Reason:  ReasonCreated,
	}
}

func (t *TriggerHandler) updateResource(ctx context.Context, rt *v1beta1.ResourceTemplate) controller.Result {
	patchValue, err := json.Marshal(rt.Spec)
	if err != nil {
		return controller.Result{
			Error:  fmt.Errorf("failed to marshal patch value: %w", err),
			Reason: ReasonUpdateFailed,
		}
	}

	patch, err := json.Marshal([]v1beta1.JSONPatch{
		{
			Operation: v1beta1.JSONPatchOpReplace,
			Path:      "/spec",
			Value:     &extv1.JSON{Raw: patchValue},
		},
	})
	if err != nil {
		return controller.Result{
			Error:  fmt.Errorf("failed to marshal resource template spec: %w", err),
			Reason: ReasonUpdateFailed,
		}
	}

	if err := t.Client.Patch(ctx, rt, client.RawPatch(types.JSONPatchType, patch)); err != nil {
		if kerrors.IsNotFound(err) {
			return controller.Result{
				Message: fmt.Sprintf("Resource template does not exist: %s", rt.Name),
				Reason:  ReasonNotExist,
			}
		}

		return controller.Result{
			Error:  fmt.Errorf("failed to patch resource template: %w", err),
			Reason: ReasonUpdateFailed,
		}
	}

	return controller.Result{
		Message: fmt.Sprintf("Updated resource template: %s", rt.Name),
		Reason:  ReasonUpdated,
	}
}

func (t *TriggerHandler) applyResource(ctx context.Context, rt *v1beta1.ResourceTemplate) controller.Result {
	result := t.createResource(ctx, rt)
	if result.Reason != ReasonAlreadyExists {
		return result
	}

	return t.updateResource(ctx, rt)
}

func (t *TriggerHandler) deleteResource(ctx context.Context, rt *v1beta1.ResourceTemplate) controller.Result {
	if err := t.Client.Delete(ctx, rt); err != nil {
		if kerrors.IsNotFound(err) {
			return controller.Result{
				Message: fmt.Sprintf("Resource template does not exist: %s", rt.Name),
				Reason:  ReasonNotExist,
			}
		}

		return controller.Result{
			Error:  fmt.Errorf("failed to delete resource template: %w", err),
			Reason: ReasonDeleteFailed,
		}
	}

	return controller.Result{
		Message: fmt.Sprintf("Deleted resource template: %s", rt.Name),
		Reason:  ReasonDeleted,
	}
}
