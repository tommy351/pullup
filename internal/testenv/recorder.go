package testenv

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Recorder interface {
	Changes() []Change
}

var _ Recorder = (*recorderWriter)(nil)

type recorderWriter struct {
	writer  client.Writer
	changes []Change
	scheme  *runtime.Scheme
}

func (r *recorderWriter) append(changeType string, obj runtime.Object) {
	gvk, err := apiutil.GVKForObject(obj, r.scheme)
	if err != nil {
		return
	}

	o, err := meta.Accessor(obj)
	if err != nil {
		return
	}

	r.changes = append(r.changes, Change{
		GroupVersionKind: gvk,
		NamespacedName: types.NamespacedName{
			Namespace: o.GetNamespace(),
			Name:      o.GetName(),
		},
		Type: changeType,
	})
}

func (r *recorderWriter) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	if err := r.writer.Create(ctx, obj, opts...); err != nil {
		return fmt.Errorf("create failed: %w", err)
	}

	r.append("create", obj)

	return nil
}

func (r *recorderWriter) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	if err := r.writer.Delete(ctx, obj, opts...); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	r.append("delete", obj)

	return nil
}

func (r *recorderWriter) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	if err := r.writer.Update(ctx, obj, opts...); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	r.append("update", obj)

	return nil
}

func (r *recorderWriter) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	if err := r.writer.Patch(ctx, obj, patch, opts...); err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}

	r.append("patch", obj)

	return nil
}

func (r *recorderWriter) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	if err := r.writer.DeleteAllOf(ctx, obj, opts...); err != nil {
		return fmt.Errorf("delete all failed: %w", err)
	}

	r.append("delete", obj)

	return nil
}

func (r *recorderWriter) Changes() []Change {
	return r.changes
}

var _ Recorder = (*recorderStatusClient)(nil)

type recorderStatusClient struct {
	writer  client.StatusWriter
	scheme  *runtime.Scheme
	changes []Change
}

func (r *recorderStatusClient) Status() client.StatusWriter {
	return r
}

func (r *recorderStatusClient) append(changeType string, obj runtime.Object) {
	gvk, err := apiutil.GVKForObject(obj, r.scheme)
	if err != nil {
		return
	}

	o, err := meta.Accessor(obj)
	if err != nil {
		return
	}

	r.changes = append(r.changes, Change{
		GroupVersionKind: gvk,
		NamespacedName: types.NamespacedName{
			Namespace: o.GetNamespace(),
			Name:      o.GetName(),
		},
		Type: changeType,
	})
}

func (r *recorderStatusClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	if err := r.writer.Update(ctx, obj, opts...); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	r.append("update", obj)

	return nil
}

func (r *recorderStatusClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	if err := r.writer.Patch(ctx, obj, patch, opts...); err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}

	r.append("patch", obj)

	return nil
}

func (r *recorderStatusClient) Changes() []Change {
	return r.changes
}
