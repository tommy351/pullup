package resourcetemplate

import (
	"fmt"

	"github.com/tommy351/pullup/internal/template"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1beta1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func renderWebhookPatch(rt *v1beta1.ResourceTemplate, patch *v1beta1.WebhookPatch) (*v1beta1.WebhookPatch, error) {
	result := *patch

	switch {
	case patch.Merge != nil && patch.Merge.Raw != nil:
		rendered, err := renderJSON(*patch.Merge, rt.Spec.Data)
		if err != nil {
			return nil, err
		}

		result.Merge = &rendered

	case len(patch.JSONPatch) > 0:
		result.JSONPatch = make([]v1beta1.JSONPatch, len(patch.JSONPatch))

		for i, p := range patch.JSONPatch {
			result.JSONPatch[i] = p

			if p.Value != nil {
				rendered, err := renderJSON(*p.Value, rt.Spec.Data)
				if err != nil {
					return nil, err
				}

				result.JSONPatch[i].Value = &rendered
			}
		}
	}

	return &result, nil
}

func renderJSON(input, data extv1.JSON) (extv1.JSON, error) {
	if input.Raw == nil || data.Raw == nil {
		return input, nil
	}

	s, err := template.RenderFromJSON(string(input.Raw), data)
	if err != nil {
		return extv1.JSON{}, fmt.Errorf("render failed: %w", err)
	}

	return extv1.JSON{Raw: []byte(s)}, nil
}
