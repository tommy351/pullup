package controller

import (
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewEventRecorder(mgr manager.Manager) record.EventRecorder {
	return mgr.GetEventRecorderFor("pullup-controller")
}

func NewClient(mgr manager.Manager) client.Client {
	return mgr.GetClient()
}

func NewAPIReader(mgr manager.Manager) client.Reader {
	return mgr.GetAPIReader()
}
