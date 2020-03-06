package driver_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/drivertest"
	"github.com/graphql-editor/stucco/pkg/driver/protohttp"
	azuredriver "github.com/graphql-editor/stucco/pkg/providers/azure/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProtobufNewClient(t *testing.T) {
	var pc azuredriver.ProtobufClient
	assert.Equal(
		t,
		&protohttp.Client{
			Client: http.DefaultClient,
			URL:    "mockurl",
		},
		pc.New("mockurl"),
	)
	pc.Client = &http.Client{}
	assert.Equal(
		t,
		&protohttp.Client{
			Client: &http.Client{},
			URL:    "mockurl",
		},
		pc.New("mockurl"),
	)
}

type mockWorkerClient struct {
	mock.Mock
}

func (m *mockWorkerClient) New(u string) driver.Driver {
	return m.Called(u).Get(0).(driver.Driver)
}

func TestDriver(t *testing.T) {
	os.Setenv("STUCCO_WORKER_BASE_URL", "http://mockurl")
	defer os.Unsetenv("STUCCO_WORKER_BASE_URL")
	var mockDriver drivertest.MockDriver
	var mockWorkerClient mockWorkerClient
	mockWorkerClient.On("New", "http://mockurl").Return(&mockDriver)
	d := azuredriver.Driver{
		WorkerClient: &mockWorkerClient,
	}

	// Test FieldResolve
	mockDriver.On("FieldResolve", driver.FieldResolveInput{}).Return(driver.FieldResolveOutput{})
	assert.Equal(t, driver.FieldResolveOutput{}, d.FieldResolve(driver.FieldResolveInput{}))
	mockDriver.AssertCalled(t, "FieldResolve", driver.FieldResolveInput{})

	// Test InterfaceResolveType
	mockDriver.On("InterfaceResolveType", driver.InterfaceResolveTypeInput{}).Return(driver.InterfaceResolveTypeOutput{})
	assert.Equal(t, driver.InterfaceResolveTypeOutput{}, d.InterfaceResolveType(driver.InterfaceResolveTypeInput{}))
	mockDriver.AssertCalled(t, "InterfaceResolveType", driver.InterfaceResolveTypeInput{})

	// Test ScalarParse
	mockDriver.On("ScalarParse", driver.ScalarParseInput{}).Return(driver.ScalarParseOutput{})
	assert.Equal(t, driver.ScalarParseOutput{}, d.ScalarParse(driver.ScalarParseInput{}))
	mockDriver.AssertCalled(t, "ScalarParse", driver.ScalarParseInput{})

	// Test ScalarSerialize
	mockDriver.On("ScalarSerialize", driver.ScalarSerializeInput{}).Return(driver.ScalarSerializeOutput{})
	assert.Equal(t, driver.ScalarSerializeOutput{}, d.ScalarSerialize(driver.ScalarSerializeInput{}))
	mockDriver.AssertCalled(t, "ScalarSerialize", driver.ScalarSerializeInput{})

	// Test UnionResolveType
	mockDriver.On("UnionResolveType", driver.UnionResolveTypeInput{}).Return(driver.UnionResolveTypeOutput{})
	assert.Equal(t, driver.UnionResolveTypeOutput{}, d.UnionResolveType(driver.UnionResolveTypeInput{}))
	mockDriver.AssertCalled(t, "UnionResolveType", driver.UnionResolveTypeInput{})

	// Test Stream
	mockDriver.On("Stream", driver.StreamInput{}).Return(driver.StreamOutput{})
	assert.Equal(t, driver.StreamOutput{}, d.Stream(driver.StreamInput{}))
	mockDriver.AssertCalled(t, "Stream", driver.StreamInput{})

	os.Setenv("STUCCO_WORKER_BASE_URL", "://mockurl")

	// Test FieldResolve
	f := d.FieldResolve(driver.FieldResolveInput{})
	assert.NotNil(t, f.Error)

	// Test InterfaceResolveType
	i := d.InterfaceResolveType(driver.InterfaceResolveTypeInput{})
	assert.NotNil(t, i.Error)

	// Test ScalarParse
	sp := d.ScalarParse(driver.ScalarParseInput{})
	assert.NotNil(t, sp.Error)

	// Test ScalarSerialize
	ss := d.ScalarSerialize(driver.ScalarSerializeInput{})
	assert.NotNil(t, ss.Error)

	// Test UnionResolveType
	u := d.UnionResolveType(driver.UnionResolveTypeInput{})
	assert.NotNil(t, u.Error)

	// Test Stream
	s := d.Stream(driver.StreamInput{})
	assert.NotNil(t, s.Error)

	mockWorkerClient.AssertNumberOfCalls(t, "New", 6)
}
