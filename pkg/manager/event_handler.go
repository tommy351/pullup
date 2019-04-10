package manager

//go:generate mockgen -source=$GOFILE -destination=event_handler_mock_test.go -package=$GOPACKAGE EventHandler

import "context"

type EventHandler interface {
	OnUpdate(ctx context.Context, obj interface{}) error
	OnDelete(ctx context.Context, obj interface{}) error
}
