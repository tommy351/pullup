package kubernetes

import (
	"bytes"
	"encoding/json"
	"html/template"

	"github.com/evanphx/json-patch"
	"github.com/tommy351/pullup/pkg/config"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type jsonPatch struct {
	Op    string      `json:"op,omitempty"`
	Path  string      `json:"path,omitempty"`
	From  string      `json:"from,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

var commonPatches = []config.ResourcePatch{
	{Remove: "/status"},
	{Replace: "/metadata/name", Value: "{{ .ModifiedName }}"},
	{Remove: "/metadata/creationTimestamp"},
	{Remove: "/metadata/resourceVersion"},
	{Remove: "/metadata/selfLink"},
	{Remove: "/metadata/uid"},
	{Remove: "/metadata/generation"},
}

var typedPatchMap = map[string]map[string][]config.ResourcePatch{
	"v1": {
		"Service": {
			{Remove: "/spec/clusterIP"},
			{Remove: "/spec/ports/nodePort"},
		},
	},
}

func PatchResource(input rest.Result, resource *Resource) ([]byte, error) {
	meta, err := getTypeMeta(input)

	if err != nil {
		return nil, err
	}

	patches := commonPatches

	if gvPatches, ok := typedPatchMap[meta.APIVersion]; ok {
		if kindPatches, ok := gvPatches[meta.Kind]; ok {
			patches = append(patches, kindPatches...)
		}
	}

	patches = append(patches, resource.Patch...)

	raw, err := input.Raw()

	if err != nil {
		return nil, err
	}

	patch, err := createJSONPatch(resource, patches)

	if err != nil {
		return nil, err
	}

	return patch.Apply(raw)
}

func getTypeMeta(input rest.Result) (*runtime.TypeMeta, error) {
	raw, err := input.Raw()

	if err != nil {
		return nil, err
	}

	var meta runtime.TypeMeta

	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func createJSONPatch(resource *Resource, resPatches []config.ResourcePatch) (jsonpatch.Patch, error) {
	var (
		patches []jsonPatch
		err     error
	)

	for _, patch := range resPatches {
		var p jsonPatch

		switch {
		case patch.Add != "":
			p.Op = "add"
			p.Path = patch.Add

			if p.Value, err = renderTemplate(patch.Value, resource); err != nil {
				return nil, err
			}

		case patch.Remove != "":
			p.Op = "remove"
			p.Path = patch.Remove

		case patch.Replace != "":
			p.Op = "replace"
			p.Path = patch.Replace

			if p.Value, err = renderTemplate(patch.Value, resource); err != nil {
				return nil, err
			}

		case patch.Copy != "":
			p.Op = "copy"
			p.From = patch.Copy
			p.Path = patch.Path

		case patch.Move != "":
			p.Op = "move"
			p.From = patch.Move
			p.Path = patch.Path

		case patch.Test != "":
			p.Op = "test"
			p.Path = patch.Test

			if p.Value, err = renderTemplate(patch.Value, resource); err != nil {
				return nil, err
			}
		}

		patches = append(patches, p)
	}

	buf, err := json.Marshal(patches)

	if err != nil {
		return nil, err
	}

	return jsonpatch.DecodePatch(buf)
}

func renderTemplate(input interface{}, data interface{}) (interface{}, error) {
	// TODO: Support other types
	text, ok := input.(string)

	if !ok {
		return input, nil
	}

	tmpl, err := template.New("").Parse(text)

	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.String(), nil
}
