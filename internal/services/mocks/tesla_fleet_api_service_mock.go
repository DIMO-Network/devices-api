// Code generated by MockGen. DO NOT EDIT.
// Source: internal/services/tesla_fleet_api_service.go
//
// Generated by this command:
//
//	mockgen -source=internal/services/tesla_fleet_api_service.go -destination=internal/services/mocks/tesla_fleet_api_service_mock.go
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	services "github.com/DIMO-Network/devices-api/internal/services"
	gomock "go.uber.org/mock/gomock"
)

// MockTeslaFleetAPIService is a mock of TeslaFleetAPIService interface.
type MockTeslaFleetAPIService struct {
	ctrl     *gomock.Controller
	recorder *MockTeslaFleetAPIServiceMockRecorder
}

// MockTeslaFleetAPIServiceMockRecorder is the mock recorder for MockTeslaFleetAPIService.
type MockTeslaFleetAPIServiceMockRecorder struct {
	mock *MockTeslaFleetAPIService
}

// NewMockTeslaFleetAPIService creates a new mock instance.
func NewMockTeslaFleetAPIService(ctrl *gomock.Controller) *MockTeslaFleetAPIService {
	mock := &MockTeslaFleetAPIService{ctrl: ctrl}
	mock.recorder = &MockTeslaFleetAPIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTeslaFleetAPIService) EXPECT() *MockTeslaFleetAPIServiceMockRecorder {
	return m.recorder
}

// CompleteTeslaAuthCodeExchange mocks base method.
func (m *MockTeslaFleetAPIService) CompleteTeslaAuthCodeExchange(ctx context.Context, authCode, redirectURI, region string) (*services.TeslaAuthCodeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompleteTeslaAuthCodeExchange", ctx, authCode, redirectURI, region)
	ret0, _ := ret[0].(*services.TeslaAuthCodeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CompleteTeslaAuthCodeExchange indicates an expected call of CompleteTeslaAuthCodeExchange.
func (mr *MockTeslaFleetAPIServiceMockRecorder) CompleteTeslaAuthCodeExchange(ctx, authCode, redirectURI, region any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompleteTeslaAuthCodeExchange", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).CompleteTeslaAuthCodeExchange), ctx, authCode, redirectURI, region)
}

// GetVehicle mocks base method.
func (m *MockTeslaFleetAPIService) GetVehicle(ctx context.Context, token, region string, vehicleID int) (*services.TeslaVehicle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVehicle", ctx, token, region, vehicleID)
	ret0, _ := ret[0].(*services.TeslaVehicle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVehicle indicates an expected call of GetVehicle.
func (mr *MockTeslaFleetAPIServiceMockRecorder) GetVehicle(ctx, token, region, vehicleID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVehicle", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).GetVehicle), ctx, token, region, vehicleID)
}

// GetVehicles mocks base method.
func (m *MockTeslaFleetAPIService) GetVehicles(ctx context.Context, token, region string) ([]services.TeslaVehicle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVehicles", ctx, token, region)
	ret0, _ := ret[0].([]services.TeslaVehicle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVehicles indicates an expected call of GetVehicles.
func (mr *MockTeslaFleetAPIServiceMockRecorder) GetVehicles(ctx, token, region any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVehicles", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).GetVehicles), ctx, token, region)
}

// WakeUpVehicle mocks base method.
func (m *MockTeslaFleetAPIService) WakeUpVehicle(ctx context.Context, token, region string, vehicleID int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WakeUpVehicle", ctx, token, region, vehicleID)
	ret0, _ := ret[0].(error)
	return ret0
}

// WakeUpVehicle indicates an expected call of WakeUpVehicle.
func (mr *MockTeslaFleetAPIServiceMockRecorder) WakeUpVehicle(ctx, token, region, vehicleID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WakeUpVehicle", reflect.TypeOf((*MockTeslaFleetAPIService)(nil).WakeUpVehicle), ctx, token, region, vehicleID)
}
