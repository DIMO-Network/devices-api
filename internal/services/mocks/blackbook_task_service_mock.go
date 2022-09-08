// Code generated by MockGen. DO NOT EDIT.
// Source: blackbook_task_service.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	services "github.com/DIMO-Network/devices-api/internal/services"
	gomock "github.com/golang/mock/gomock"
)

// MockBlackbookTaskService is a mock of BlackbookTaskService interface.
type MockBlackbookTaskService struct {
	ctrl     *gomock.Controller
	recorder *MockBlackbookTaskServiceMockRecorder
}

// MockBlackbookTaskServiceMockRecorder is the mock recorder for MockBlackbookTaskService.
type MockBlackbookTaskServiceMockRecorder struct {
	mock *MockBlackbookTaskService
}

// NewMockBlackbookTaskService creates a new mock instance.
func NewMockBlackbookTaskService(ctrl *gomock.Controller) *MockBlackbookTaskService {
	mock := &MockBlackbookTaskService{ctrl: ctrl}
	mock.recorder = &MockBlackbookTaskServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBlackbookTaskService) EXPECT() *MockBlackbookTaskServiceMockRecorder {
	return m.recorder
}

// GetTaskStatus mocks base method.
func (m *MockBlackbookTaskService) GetTaskStatus(ctx context.Context, taskID string) (*services.BlackbookTask, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskStatus", ctx, taskID)
	ret0, _ := ret[0].(*services.BlackbookTask)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTaskStatus indicates an expected call of GetTaskStatus.
func (mr *MockBlackbookTaskServiceMockRecorder) GetTaskStatus(ctx, taskID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskStatus", reflect.TypeOf((*MockBlackbookTaskService)(nil).GetTaskStatus), ctx, taskID)
}

// StartBlackbookUpdate mocks base method.
func (m *MockBlackbookTaskService) StartBlackbookUpdate(deviceDefinitionID, userDeviceID, vin string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartBlackbookUpdate", deviceDefinitionID, userDeviceID, vin)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartBlackbookUpdate indicates an expected call of StartBlackbookUpdate.
func (mr *MockBlackbookTaskServiceMockRecorder) StartBlackbookUpdate(deviceDefinitionID, userDeviceID, vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartBlackbookUpdate", reflect.TypeOf((*MockBlackbookTaskService)(nil).StartBlackbookUpdate), deviceDefinitionID, userDeviceID, vin)
}

// StartConsumer mocks base method.
func (m *MockBlackbookTaskService) StartConsumer(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartConsumer", ctx)
}

// StartConsumer indicates an expected call of StartConsumer.
func (mr *MockBlackbookTaskServiceMockRecorder) StartConsumer(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartConsumer", reflect.TypeOf((*MockBlackbookTaskService)(nil).StartConsumer), ctx)
}
