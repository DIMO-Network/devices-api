// Code generated by MockGen. DO NOT EDIT.
// Source: drivly_api_service.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	reflect "reflect"

	services "github.com/DIMO-Network/devices-api/internal/services"
	gomock "github.com/golang/mock/gomock"
)

// MockDrivlyAPIService is a mock of DrivlyAPIService interface.
type MockDrivlyAPIService struct {
	ctrl     *gomock.Controller
	recorder *MockDrivlyAPIServiceMockRecorder
}

// MockDrivlyAPIServiceMockRecorder is the mock recorder for MockDrivlyAPIService.
type MockDrivlyAPIServiceMockRecorder struct {
	mock *MockDrivlyAPIService
}

// NewMockDrivlyAPIService creates a new mock instance.
func NewMockDrivlyAPIService(ctrl *gomock.Controller) *MockDrivlyAPIService {
	mock := &MockDrivlyAPIService{ctrl: ctrl}
	mock.recorder = &MockDrivlyAPIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDrivlyAPIService) EXPECT() *MockDrivlyAPIServiceMockRecorder {
	return m.recorder
}

// GetAutocheckByVIN mocks base method.
func (m *MockDrivlyAPIService) GetAutocheckByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAutocheckByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAutocheckByVIN indicates an expected call of GetAutocheckByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetAutocheckByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAutocheckByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetAutocheckByVIN), vin)
}

// GetBuildByVIN mocks base method.
func (m *MockDrivlyAPIService) GetBuildByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBuildByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBuildByVIN indicates an expected call of GetBuildByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetBuildByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBuildByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetBuildByVIN), vin)
}

// GetCargurusByVIN mocks base method.
func (m *MockDrivlyAPIService) GetCargurusByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCargurusByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCargurusByVIN indicates an expected call of GetCargurusByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetCargurusByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCargurusByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetCargurusByVIN), vin)
}

// GetCarmaxByVIN mocks base method.
func (m *MockDrivlyAPIService) GetCarmaxByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCarmaxByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCarmaxByVIN indicates an expected call of GetCarmaxByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetCarmaxByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCarmaxByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetCarmaxByVIN), vin)
}

// GetCarstoryByVIN mocks base method.
func (m *MockDrivlyAPIService) GetCarstoryByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCarstoryByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCarstoryByVIN indicates an expected call of GetCarstoryByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetCarstoryByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCarstoryByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetCarstoryByVIN), vin)
}

// GetCarvanaByVIN mocks base method.
func (m *MockDrivlyAPIService) GetCarvanaByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCarvanaByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCarvanaByVIN indicates an expected call of GetCarvanaByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetCarvanaByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCarvanaByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetCarvanaByVIN), vin)
}

// GetEdmundsByVIN mocks base method.
func (m *MockDrivlyAPIService) GetEdmundsByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEdmundsByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEdmundsByVIN indicates an expected call of GetEdmundsByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetEdmundsByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEdmundsByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetEdmundsByVIN), vin)
}

// GetExtendedOffersByVIN mocks base method.
func (m *MockDrivlyAPIService) GetExtendedOffersByVIN(vin string) (*services.DrivlyVINSummary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExtendedOffersByVIN", vin)
	ret0, _ := ret[0].(*services.DrivlyVINSummary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExtendedOffersByVIN indicates an expected call of GetExtendedOffersByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetExtendedOffersByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExtendedOffersByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetExtendedOffersByVIN), vin)
}

// GetKBBByVIN mocks base method.
func (m *MockDrivlyAPIService) GetKBBByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKBBByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKBBByVIN indicates an expected call of GetKBBByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetKBBByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKBBByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetKBBByVIN), vin)
}

// GetOffersByVIN mocks base method.
func (m *MockDrivlyAPIService) GetOffersByVIN(vin string, mileage float64, zipcode string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOffersByVIN", vin, mileage, zipcode)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOffersByVIN indicates an expected call of GetOffersByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetOffersByVIN(vin, mileage, zipcode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOffersByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetOffersByVIN), vin, mileage, zipcode)
}

// GetTMVByVIN mocks base method.
func (m *MockDrivlyAPIService) GetTMVByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTMVByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTMVByVIN indicates an expected call of GetTMVByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetTMVByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTMVByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetTMVByVIN), vin)
}

// GetVINInfo mocks base method.
func (m *MockDrivlyAPIService) GetVINInfo(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVINInfo", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVINInfo indicates an expected call of GetVINInfo.
func (mr *MockDrivlyAPIServiceMockRecorder) GetVINInfo(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVINInfo", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetVINInfo), vin)
}

// GetVINPricing mocks base method.
func (m *MockDrivlyAPIService) GetVINPricing(vin string, mileage float64, zipcode string) (map[string]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVINPricing", vin, mileage, zipcode)
	ret0, _ := ret[0].(map[string]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVINPricing indicates an expected call of GetVINPricing.
func (mr *MockDrivlyAPIServiceMockRecorder) GetVINPricing(vin, mileage, zipcode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVINPricing", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetVINPricing), vin, mileage, zipcode)
}

// GetVRoomByVIN mocks base method.
func (m *MockDrivlyAPIService) GetVRoomByVIN(vin string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVRoomByVIN", vin)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVRoomByVIN indicates an expected call of GetVRoomByVIN.
func (mr *MockDrivlyAPIServiceMockRecorder) GetVRoomByVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVRoomByVIN", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetVRoomByVIN), vin)
}
