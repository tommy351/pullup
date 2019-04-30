package golden

import (
	"time"

	"github.com/tommy351/pullup/internal/testutil"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// nolint: gochecknoglobals
var zeroUID = types.UID("00000000-0000-0000-0000-000000000000")

func PrepareObjects(input []runtime.Object) []runtime.Object {
	output := testutil.MapObjects(input, func(object runtime.Object) {
		o, err := meta.Accessor(object)

		if err == nil {
			o.SetUID(zeroUID)
			o.SetCreationTimestamp(metav1.Time{Time: time.Unix(0, 0)})
			o.SetResourceVersion("0")
			o.SetGeneration(0)
			o.SetClusterName("")
			o.SetSelfLink("")
		}
	})

	testutil.SortObjects(output)
	return output
}
