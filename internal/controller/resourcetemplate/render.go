package resourcetemplate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tommy351/pullup/internal/jsonutil"
	"github.com/tommy351/pullup/internal/template"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) renderTriggerPatches(ctx context.Context, rt *v1beta1.ResourceTemplate) ([]v1beta1.TriggerPatch, error) {
	output := make([]v1beta1.TriggerPatch, len(rt.Spec.Patches))
	raw := rt.Spec.Data.Raw
	if raw == nil {
		raw = []byte("{}")
	}

	var (
		err  error
		data interface{}
	)

	addedKeys := map[string]interface{}{
		v1beta1.DataKeyResource: rt,
	}

	if ref := rt.Spec.TriggerRef; ref != nil {
		webhook, err := r.getObject(ctx, ref.GroupVersionKind(), types.NamespacedName{
			Namespace: rt.Namespace,
			Name:      ref.Name,
		})
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}

		if webhook != nil {
			addedKeys[v1beta1.DataKeyTrigger] = webhook
		}
	}

	raw, err = jsonutil.AddMapKeys(raw, addedKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to mutate data: %w", err)
	}

	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	for i, patch := range rt.Spec.Patches {
		patch := patch
		result, err := r.renderTriggerPatch(rt, &patch, data)
		if err != nil {
			return nil, err
		}

		output[i] = *result
	}

	return output, nil
}

func (r *Reconciler) renderTriggerPatch(rt *v1beta1.ResourceTemplate, patch *v1beta1.TriggerPatch, data interface{}) (*v1beta1.TriggerPatch, error) {
	var err error
	result := patch.DeepCopy()

	if result.APIVersion, err = template.Render(patch.APIVersion, data); err != nil {
		return nil, fmt.Errorf("failed to render apiVersion: %w", err)
	}

	if result.Kind, err = template.Render(patch.Kind, data); err != nil {
		return nil, fmt.Errorf("failed to render kind: %w", err)
	}

	if result.SourceName, err = template.Render(patch.SourceName, data); err != nil {
		return nil, fmt.Errorf("failed to render sourceName: %w", err)
	}

	if result.TargetName, err = template.Render(patch.TargetName, data); err != nil {
		return nil, fmt.Errorf("failed to render targetName: %w", err)
	}

	if result.TargetName == "" {
		result.TargetName = rt.Name
	}

	switch {
	case patch.Merge != nil && patch.Merge.Raw != nil:
		rendered, err := template.Render(string(patch.Merge.Raw), data)
		if err != nil {
			return nil, fmt.Errorf("failed to render merge: %w", err)
		}

		result.Merge = &extv1.JSON{Raw: []byte(rendered)}

	case len(patch.JSONPatch) > 0:
		result.JSONPatch = make([]v1beta1.JSONPatch, len(patch.JSONPatch))

		for i, p := range patch.JSONPatch {
			result.JSONPatch[i] = p

			if p.Value != nil {
				rendered, err := template.Render(string(p.Value.Raw), data)
				if err != nil {
					return nil, fmt.Errorf("failed to render jsonPatch: %w", err)
				}

				result.JSONPatch[i].Value = &extv1.JSON{Raw: []byte(rendered)}
			}
		}
	}

	return result, nil
}
