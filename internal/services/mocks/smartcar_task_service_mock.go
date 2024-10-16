// Code generated by MockGen. DO NOT EDIT.
// Source: smartcar_task_service.go
//
// Generated by this command:
//
//	mockgen -source smartcar_task_service.go -destination mocks/smartcar_task_service_mock.go
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	reflect "reflect"

	models "github.com/DIMO-Network/devices-api/models"
	gomock "go.uber.org/mock/gomock"
)

// MockSmartcarTaskService is a mock of SmartcarTaskService interface.
type MockSmartcarTaskService struct {
	ctrl     *gomock.Controller
	recorder *MockSmartcarTaskServiceMockRecorder
}

// MockSmartcarTaskServiceMockRecorder is the mock recorder for MockSmartcarTaskService.
type MockSmartcarTaskServiceMockRecorder struct {
	mock *MockSmartcarTaskService
}

// NewMockSmartcarTaskService creates a new mock instance.
func NewMockSmartcarTaskService(ctrl *gomock.Controller) *MockSmartcarTaskService {
	mock := &MockSmartcarTaskService{ctrl: ctrl}
	mock.recorder = &MockSmartcarTaskServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSmartcarTaskService) EXPECT() *MockSmartcarTaskServiceMockRecorder {
	return m.recorder
}

// LockDoors mocks base method.
func (m *MockSmartcarTaskService) LockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LockDoors", udai)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LockDoors indicates an expected call of LockDoors.
func (mr *MockSmartcarTaskServiceMockRecorder) LockDoors(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LockDoors", reflect.TypeOf((*MockSmartcarTaskService)(nil).LockDoors), udai)
}

// Refresh mocks base method.
func (m *MockSmartcarTaskService) Refresh(udai *models.UserDeviceAPIIntegration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refresh", udai)
	ret0, _ := ret[0].(error)
	return ret0
}

// Refresh indicates an expected call of Refresh.
func (mr *MockSmartcarTaskServiceMockRecorder) Refresh(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockSmartcarTaskService)(nil).Refresh), udai)
}

// StartPoll mocks base method.
func (m *MockSmartcarTaskService) StartPoll(udai *models.UserDeviceAPIIntegration, sd *models.SyntheticDevice) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartPoll", udai, sd)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartPoll indicates an expected call of StartPoll.
func (mr *MockSmartcarTaskServiceMockRecorder) StartPoll(udai, sd any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartPoll", reflect.TypeOf((*MockSmartcarTaskService)(nil).StartPoll), udai, sd)
}

// StopPoll mocks base method.
func (m *MockSmartcarTaskService) StopPoll(udai *models.UserDeviceAPIIntegration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopPoll", udai)
	ret0, _ := ret[0].(error)
	return ret0
}

// StopPoll indicates an expected call of StopPoll.
func (mr *MockSmartcarTaskServiceMockRecorder) StopPoll(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopPoll", reflect.TypeOf((*MockSmartcarTaskService)(nil).StopPoll), udai)
}

// UnlockDoors mocks base method.
func (m *MockSmartcarTaskService) UnlockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnlockDoors", udai)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UnlockDoors indicates an expected call of UnlockDoors.
func (mr *MockSmartcarTaskServiceMockRecorder) UnlockDoors(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnlockDoors", reflect.TypeOf((*MockSmartcarTaskService)(nil).UnlockDoors), udai)
}
