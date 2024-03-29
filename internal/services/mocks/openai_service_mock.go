// Code generated by MockGen. DO NOT EDIT.
// Source: internal/services/openai.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	reflect "reflect"

	services "github.com/DIMO-Network/devices-api/internal/services"
	gomock "go.uber.org/mock/gomock"
)

// MockOpenAI is a mock of OpenAI interface.
type MockOpenAI struct {
	ctrl     *gomock.Controller
	recorder *MockOpenAIMockRecorder
}

// MockOpenAIMockRecorder is the mock recorder for MockOpenAI.
type MockOpenAIMockRecorder struct {
	mock *MockOpenAI
}

// NewMockOpenAI creates a new mock instance.
func NewMockOpenAI(ctrl *gomock.Controller) *MockOpenAI {
	mock := &MockOpenAI{ctrl: ctrl}
	mock.recorder = &MockOpenAIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOpenAI) EXPECT() *MockOpenAIMockRecorder {
	return m.recorder
}

// GetErrorCodesDescription mocks base method.
func (m *MockOpenAI) GetErrorCodesDescription(make, model string, errorCodes []string) ([]services.ErrorCodesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetErrorCodesDescription", make, model, errorCodes)
	ret0, _ := ret[0].([]services.ErrorCodesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetErrorCodesDescription indicates an expected call of GetErrorCodesDescription.
func (mr *MockOpenAIMockRecorder) GetErrorCodesDescription(make, model, errorCodes interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetErrorCodesDescription", reflect.TypeOf((*MockOpenAI)(nil).GetErrorCodesDescription), make, model, errorCodes)
}
