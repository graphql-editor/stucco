package drivertest

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/stretchr/testify/mock"
)

// MockDriver is a mock interface for testing
type MockDriver struct {
	mock.Mock
}

func (m *MockDriver) SetSecrets(in driver.SetSecretsInput) driver.SetSecretsOutput {
	return m.Called(in).Get(0).(driver.SetSecretsOutput)
}
func (m *MockDriver) FieldResolve(in driver.FieldResolveInput) driver.FieldResolveOutput {
	return m.Called(in).Get(0).(driver.FieldResolveOutput)
}
func (m *MockDriver) InterfaceResolveType(in driver.InterfaceResolveTypeInput) driver.InterfaceResolveTypeOutput {
	return m.Called(in).Get(0).(driver.InterfaceResolveTypeOutput)
}
func (m *MockDriver) ScalarParse(in driver.ScalarParseInput) driver.ScalarParseOutput {
	return m.Called(in).Get(0).(driver.ScalarParseOutput)
}
func (m *MockDriver) ScalarSerialize(in driver.ScalarSerializeInput) driver.ScalarSerializeOutput {
	return m.Called(in).Get(0).(driver.ScalarSerializeOutput)
}
func (m *MockDriver) UnionResolveType(in driver.UnionResolveTypeInput) driver.UnionResolveTypeOutput {
	return m.Called(in).Get(0).(driver.UnionResolveTypeOutput)
}
func (m *MockDriver) Stream(in driver.StreamInput) driver.StreamOutput {
	return m.Called(in).Get(0).(driver.StreamOutput)
}

func (m *MockDriver) SubscriptionListen(in driver.SubscriptionListenInput) driver.SubscriptionListenOutput {
	return m.Called(in).Get(0).(driver.SubscriptionListenOutput)
}

func (m *MockDriver) SubscriptionConnection(in driver.SubscriptionConnectionInput) driver.SubscriptionConnectionOutput {
	return m.Called(in).Get(0).(driver.SubscriptionConnectionOutput)
}
