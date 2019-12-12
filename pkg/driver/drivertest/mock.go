package drivertest

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/stretchr/testify/mock"
)

// MockDriver is a mock interface for testing
type MockDriver struct {
	mock.Mock
}

func (m *MockDriver) SetSecrets(in driver.SetSecretsInput) (driver.SetSecretsOutput, error) {
	called := m.Called(in)
	return called.Get(0).(driver.SetSecretsOutput), called.Error(1)
}
func (m *MockDriver) FieldResolve(in driver.FieldResolveInput) (driver.FieldResolveOutput, error) {
	called := m.Called(in)
	return called.Get(0).(driver.FieldResolveOutput), called.Error(1)
}
func (m *MockDriver) InterfaceResolveType(in driver.InterfaceResolveTypeInput) (driver.InterfaceResolveTypeOutput, error) {
	called := m.Called(in)
	return called.Get(0).(driver.InterfaceResolveTypeOutput), called.Error(1)
}
func (m *MockDriver) ScalarParse(in driver.ScalarParseInput) (driver.ScalarParseOutput, error) {
	called := m.Called(in)
	return called.Get(0).(driver.ScalarParseOutput), called.Error(1)
}
func (m *MockDriver) ScalarSerialize(in driver.ScalarSerializeInput) (driver.ScalarSerializeOutput, error) {
	called := m.Called(in)
	return called.Get(0).(driver.ScalarSerializeOutput), called.Error(1)
}
func (m *MockDriver) UnionResolveType(in driver.UnionResolveTypeInput) (driver.UnionResolveTypeOutput, error) {
	called := m.Called(in)
	return called.Get(0).(driver.UnionResolveTypeOutput), called.Error(1)
}
func (m *MockDriver) Stream(in driver.StreamInput) (driver.StreamOutput, error) {
	called := m.Called(in)
	return called.Get(0).(driver.StreamOutput), called.Error(1)
}
