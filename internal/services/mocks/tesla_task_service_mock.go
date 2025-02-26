// Code generated by MockGen. DO NOT EDIT.
// Source: tesla_task_service.go
//
// Generated by this command:
//
//	mockgen -source tesla_task_service.go -destination mocks/tesla_task_service_mock.go
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	reflect "reflect"

	models "github.com/DIMO-Network/devices-api/models"
	gomock "go.uber.org/mock/gomock"
)

// MockTeslaTaskService is a mock of TeslaTaskService interface.
type MockTeslaTaskService struct {
	ctrl     *gomock.Controller
	recorder *MockTeslaTaskServiceMockRecorder
	isgomock struct{}
}

// MockTeslaTaskServiceMockRecorder is the mock recorder for MockTeslaTaskService.
type MockTeslaTaskServiceMockRecorder struct {
	mock *MockTeslaTaskService
}

// NewMockTeslaTaskService creates a new mock instance.
func NewMockTeslaTaskService(ctrl *gomock.Controller) *MockTeslaTaskService {
	mock := &MockTeslaTaskService{ctrl: ctrl}
	mock.recorder = &MockTeslaTaskServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTeslaTaskService) EXPECT() *MockTeslaTaskServiceMockRecorder {
	return m.recorder
}

// LockDoors mocks base method.
func (m *MockTeslaTaskService) LockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LockDoors", udai)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LockDoors indicates an expected call of LockDoors.
func (mr *MockTeslaTaskServiceMockRecorder) LockDoors(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LockDoors", reflect.TypeOf((*MockTeslaTaskService)(nil).LockDoors), udai)
}

// OpenFrunk mocks base method.
func (m *MockTeslaTaskService) OpenFrunk(udai *models.UserDeviceAPIIntegration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenFrunk", udai)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenFrunk indicates an expected call of OpenFrunk.
func (mr *MockTeslaTaskServiceMockRecorder) OpenFrunk(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenFrunk", reflect.TypeOf((*MockTeslaTaskService)(nil).OpenFrunk), udai)
}

// OpenTrunk mocks base method.
func (m *MockTeslaTaskService) OpenTrunk(udai *models.UserDeviceAPIIntegration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenTrunk", udai)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenTrunk indicates an expected call of OpenTrunk.
func (mr *MockTeslaTaskServiceMockRecorder) OpenTrunk(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenTrunk", reflect.TypeOf((*MockTeslaTaskService)(nil).OpenTrunk), udai)
}

// StartPoll mocks base method.
func (m *MockTeslaTaskService) StartPoll(udai *models.UserDeviceAPIIntegration, sd *models.SyntheticDevice) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartPoll", udai, sd)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartPoll indicates an expected call of StartPoll.
func (mr *MockTeslaTaskServiceMockRecorder) StartPoll(udai, sd any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartPoll", reflect.TypeOf((*MockTeslaTaskService)(nil).StartPoll), udai, sd)
}

// StopPoll mocks base method.
func (m *MockTeslaTaskService) StopPoll(udai *models.UserDeviceAPIIntegration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopPoll", udai)
	ret0, _ := ret[0].(error)
	return ret0
}

// StopPoll indicates an expected call of StopPoll.
func (mr *MockTeslaTaskServiceMockRecorder) StopPoll(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopPoll", reflect.TypeOf((*MockTeslaTaskService)(nil).StopPoll), udai)
}

// UnlockDoors mocks base method.
func (m *MockTeslaTaskService) UnlockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnlockDoors", udai)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UnlockDoors indicates an expected call of UnlockDoors.
func (mr *MockTeslaTaskServiceMockRecorder) UnlockDoors(udai any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnlockDoors", reflect.TypeOf((*MockTeslaTaskService)(nil).UnlockDoors), udai)
}
