package controller

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

type Result struct {
	Error     error
	EventType string
	Reason    string
	Requeue   bool
	Message   string
}

func (r Result) GetEventType() string {
	if event := r.EventType; event != "" {
		return event
	}

	if r.Error != nil {
		return corev1.EventTypeWarning
	}

	return corev1.EventTypeNormal
}

func (r Result) GetMessage() string {
	if msg := r.Message; msg != "" {
		return msg
	}

	if err := r.Error; err != nil {
		return err.Error()
	}

	return ""
}

func (r Result) RecordEvent(recorder record.EventRecorder, object runtime.Object) {
	recorder.Event(object, r.GetEventType(), r.Reason, r.GetMessage())
}
