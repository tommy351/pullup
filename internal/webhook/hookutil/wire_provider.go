package hookutil

import (
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewEventRecorder(mgr manager.Manager) record.EventRecorder {
	return mgr.GetEventRecorderFor("pullup-webhook")
}

func NewFieldIndexer(mgr manager.Manager) client.FieldIndexer {
	return mgr.GetFieldIndexer()
}
