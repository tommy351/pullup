package testenv

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EventData struct {
	Type    string
	Reason  string
	Message string
}

func ListEvents(options ...client.ListOption) ([]corev1.Event, error) {
	list := new(corev1.EventList)

	if err := GetClient().List(context.Background(), list, options...); err != nil {
		return nil, err
	}

	return list.Items, nil
}

func MapEventData(events []corev1.Event) []EventData {
	output := make([]EventData, len(events))

	for i, event := range events {
		output[i] = EventData{
			Type:    event.Type,
			Reason:  event.Reason,
			Message: event.Message,
		}
	}

	return output
}
