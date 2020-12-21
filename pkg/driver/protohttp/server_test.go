package protohttp_test

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/stretchr/testify/mock"
)

type mockMuxer struct {
	mock.Mock
}

func (m *mockMuxer) FieldResolve(in driver.FieldResolveInput) (interface{}, error) {
	called := m.Called(in)
	return called.Get(0), called.Error(1)
}

func (m *mockMuxer) InterfaceResolveType(in driver.InterfaceResolveTypeInput) (string, error) {
	called := m.Called(in)
	return called.String(0), called.Error(1)
}

func (m *mockMuxer) SetSecrets(in driver.SetSecretsInput) error {
	return m.Called(in).Error(0)
}

func (m *mockMuxer) ScalarParse(in driver.ScalarParseInput) (interface{}, error) {
	called := m.Called(in)
	return called.Get(0), called.Error(1)
}

func (m *mockMuxer) ScalarSerialize(in driver.ScalarSerializeInput) (interface{}, error) {
	called := m.Called(in)
	return called.Get(0), called.Error(1)
}

func (m *mockMuxer) UnionResolveType(in driver.UnionResolveTypeInput) (string, error) {
	called := m.Called(in)
	return called.String(0), called.Error(1)
}

func (m *mockMuxer) SubscriptionConnection(in driver.SubscriptionConnectionInput) (interface{}, error) {
	called := m.Called(in)
	return called.Get(0), called.Error(1)
}
