package reducer

import (
	"encoding/json"

	"github.com/ansel1/merry"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/model"
)

type JSONPatch struct {
	Op    string      `json:"op,omitempty"`
	Path  string      `json:"path,omitempty"`
	From  string      `json:"from,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type jsonPatchReducer jsonpatch.Patch

func (j jsonPatchReducer) Reduce(resource *model.Resource) error {
	return ReduceBytes(resource, func(data []byte, resource *model.Resource) ([]byte, error) {
		return jsonpatch.Patch(j).Apply(data)
	})
}

func NewJSONPatch(patches []JSONPatch) (Reducer, error) {
	buf, err := json.Marshal(patches)

	if err != nil {
		return nil, merry.Wrap(err)
	}

	patch, err := jsonpatch.DecodePatch(buf)

	if err != nil {
		return nil, merry.Wrap(err)
	}

	return jsonPatchReducer(patch), nil
}

func JSONPatchFromConfig(conf []config.ResourcePatch) []JSONPatch {
	var patches []JSONPatch

	for _, p := range conf {
		switch {
		case p.Add != "":
			patches = append(patches, JSONPatch{Op: "add", Path: p.Add, Value: p.Value})
		case p.Remove != "":
			patches = append(patches, JSONPatch{Op: "remove", Path: p.Remove})
		case p.Replace != "":
			patches = append(patches, JSONPatch{Op: "replace", Path: p.Replace, Value: p.Value})
		case p.Copy != "":
			patches = append(patches, JSONPatch{Op: "copy", From: p.Copy, Path: p.Path})
		case p.Move != "":
			patches = append(patches, JSONPatch{Op: "move", From: p.Move, Path: p.Path})
		case p.Test != "":
			patches = append(patches, JSONPatch{Op: "test", Path: p.Test, Value: p.Value})
		}
	}

	return patches
}
