// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DIMO-Network/devices-api/internal/controllers (interfaces: CredStore)
//
// Generated by this command:
//
//	mockgen -destination=cred_store_mock_test.go -package controllers . CredStore
//

// Package controllers is a generated GoMock package.
package controllers

import (
	context "context"
	reflect "reflect"

	tmpcred "github.com/DIMO-Network/devices-api/internal/services/tmpcred"
	common "github.com/ethereum/go-ethereum/common"
	gomock "go.uber.org/mock/gomock"
)

// MockCredStore is a mock of CredStore interface.
type MockCredStore struct {
	ctrl     *gomock.Controller
	recorder *MockCredStoreMockRecorder
}

// MockCredStoreMockRecorder is the mock recorder for MockCredStore.
type MockCredStoreMockRecorder struct {
	mock *MockCredStore
}

// NewMockCredStore creates a new mock instance.
func NewMockCredStore(ctrl *gomock.Controller) *MockCredStore {
	mock := &MockCredStore{ctrl: ctrl}
	mock.recorder = &MockCredStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCredStore) EXPECT() *MockCredStoreMockRecorder {
	return m.recorder
}

// Store mocks base method.
func (m *MockCredStore) Store(arg0 context.Context, arg1 common.Address, arg2 *tmpcred.Credential) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockCredStoreMockRecorder) Store(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockCredStore)(nil).Store), arg0, arg1, arg2)
}
