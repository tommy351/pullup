package testutil

import (
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
)

type ObjectSlice []runtime.Object

func (o ObjectSlice) Len() int {
	return len(o)
}

func (o ObjectSlice) Less(i, j int) bool {
	gvkI := o[i].GetObjectKind().GroupVersionKind()
	gvkJ := o[j].GetObjectKind().GroupVersionKind()
	nameI := o[i].(nameGetter)
	nameJ := o[j].(nameGetter)

	result := CompareMultiStrings(
		gvkI.Group, gvkJ.Group,
		gvkI.Version, gvkJ.Version,
		gvkI.Kind, gvkJ.Kind,
		nameI.GetNamespace(), nameJ.GetNamespace(),
		nameI.GetName(), nameJ.GetName(),
	)

	return result < 0
}

func (o ObjectSlice) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func SortObjects(objects []runtime.Object) {
	sort.Stable(ObjectSlice(objects))
}

func CompareMultiStrings(strs ...string) int {
	for i := 0; i < len(strs); i += 2 {
		if result := strings.Compare(strs[i], strs[i+1]); result != 0 {
			return result
		}
	}

	return 0
}
