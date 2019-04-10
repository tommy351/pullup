// Code generated by MockGen. DO NOT EDIT.
// Source: event_handler.go

// Package manager is a generated GoMock package.
package manager

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockEventHandler is a mock of EventHandler interface
type MockEventHandler struct {
	ctrl     *gomock.Controller
	recorder *MockEventHandlerMockRecorder
}

// MockEventHandlerMockRecorder is the mock recorder for MockEventHandler
type MockEventHandlerMockRecorder struct {
	mock *MockEventHandler
}

// NewMockEventHandler creates a new mock instance
func NewMockEventHandler(ctrl *gomock.Controller) *MockEventHandler {
	mock := &MockEventHandler{ctrl: ctrl}
	mock.recorder = &MockEventHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEventHandler) EXPECT() *MockEventHandlerMockRecorder {
	return m.recorder
}

// OnUpdate mocks base method
func (m *MockEventHandler) OnUpdate(ctx context.Context, obj interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OnUpdate", ctx, obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// OnUpdate indicates an expected call of OnUpdate
func (mr *MockEventHandlerMockRecorder) OnUpdate(ctx, obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnUpdate", reflect.TypeOf((*MockEventHandler)(nil).OnUpdate), ctx, obj)
}

// OnDelete mocks base method
func (m *MockEventHandler) OnDelete(ctx context.Context, obj interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OnDelete", ctx, obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// OnDelete indicates an expected call of OnDelete
func (mr *MockEventHandlerMockRecorder) OnDelete(ctx, obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnDelete", reflect.TypeOf((*MockEventHandler)(nil).OnDelete), ctx, obj)
}
