package resourcetemplate

import (
	"encoding/json"

	"github.com/tommy351/pullup/internal/template"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func renderWebhookPatch(rt *v1beta1.ResourceTemplate, patch *v1beta1.WebhookPatch) (*v1beta1.WebhookPatch, error) {
	var (
		err  error
		data interface{}
	)
	result := patch.DeepCopy()

	if raw := rt.Spec.Data.Raw; raw != nil {
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, err
		}
	}

	if result.APIVersion, err = template.Render(patch.APIVersion, data); err != nil {
		return nil, err
	}

	if result.Kind, err = template.Render(patch.Kind, data); err != nil {
		return nil, err
	}

	if result.Name, err = template.Render(patch.Name, data); err != nil {
		return nil, err
	}

	switch {
	case patch.Merge != nil && patch.Merge.Raw != nil:
		rendered, err := template.Render(string(patch.Merge.Raw), data)
		if err != nil {
			return nil, err
		}

		result.Merge = &extv1.JSON{Raw: []byte(rendered)}

	case len(patch.JSONPatch) > 0:
		result.JSONPatch = make([]v1beta1.JSONPatch, len(patch.JSONPatch))

		for i, p := range patch.JSONPatch {
			result.JSONPatch[i] = p

			if p.Value != nil {
				rendered, err := template.Render(string(p.Value.Raw), data)
				if err != nil {
					return nil, err
				}

				result.JSONPatch[i].Value = &extv1.JSON{Raw: []byte(rendered)}
			}
		}
	}

	return result, nil
}
