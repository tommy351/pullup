/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeResourceSets implements ResourceSetInterface
type FakeResourceSets struct {
	Fake *FakePullupV1alpha1
	ns   string
}

var resourcesetsResource = schema.GroupVersionResource{Group: "pullup.dev", Version: "v1alpha1", Resource: "resourcesets"}

var resourcesetsKind = schema.GroupVersionKind{Group: "pullup.dev", Version: "v1alpha1", Kind: "ResourceSet"}

// Get takes name of the resourceSet, and returns the corresponding resourceSet object, and an error if there is any.
func (c *FakeResourceSets) Get(name string, options v1.GetOptions) (result *v1alpha1.ResourceSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(resourcesetsResource, c.ns, name), &v1alpha1.ResourceSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ResourceSet), err
}

// List takes label and field selectors, and returns the list of ResourceSets that match those selectors.
func (c *FakeResourceSets) List(opts v1.ListOptions) (result *v1alpha1.ResourceSetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(resourcesetsResource, resourcesetsKind, c.ns, opts), &v1alpha1.ResourceSetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ResourceSetList{ListMeta: obj.(*v1alpha1.ResourceSetList).ListMeta}
	for _, item := range obj.(*v1alpha1.ResourceSetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested resourceSets.
func (c *FakeResourceSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(resourcesetsResource, c.ns, opts))

}

// Create takes the representation of a resourceSet and creates it.  Returns the server's representation of the resourceSet, and an error, if there is any.
func (c *FakeResourceSets) Create(resourceSet *v1alpha1.ResourceSet) (result *v1alpha1.ResourceSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(resourcesetsResource, c.ns, resourceSet), &v1alpha1.ResourceSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ResourceSet), err
}

// Update takes the representation of a resourceSet and updates it. Returns the server's representation of the resourceSet, and an error, if there is any.
func (c *FakeResourceSets) Update(resourceSet *v1alpha1.ResourceSet) (result *v1alpha1.ResourceSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(resourcesetsResource, c.ns, resourceSet), &v1alpha1.ResourceSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ResourceSet), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeResourceSets) UpdateStatus(resourceSet *v1alpha1.ResourceSet) (*v1alpha1.ResourceSet, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(resourcesetsResource, "status", c.ns, resourceSet), &v1alpha1.ResourceSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ResourceSet), err
}

// Delete takes name of the resourceSet and deletes it. Returns an error if one occurs.
func (c *FakeResourceSets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(resourcesetsResource, c.ns, name), &v1alpha1.ResourceSet{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeResourceSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(resourcesetsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ResourceSetList{})
	return err
}

// Patch applies the patch and returns the patched resourceSet.
func (c *FakeResourceSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ResourceSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(resourcesetsResource, c.ns, name, pt, data, subresources...), &v1alpha1.ResourceSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ResourceSet), err
}
