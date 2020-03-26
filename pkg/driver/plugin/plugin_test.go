package plugin_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/driver/drivertest"
	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type execCommandMock struct {
	mock.Mock
}

func (m *execCommandMock) Command(command string, args ...string) *exec.Cmd {
	i := []interface{}{command}
	for _, arg := range args {
		i = append(i, arg)
	}
	return m.Called(i...).Get(0).(func(string, ...string) *exec.Cmd)(command, args...)
}

type newPluginClientMock struct {
	mock.Mock
}

func (m *newPluginClientMock) NewPlugin(cfg *goplugin.ClientConfig) plugin.Client {
	return m.Called(cfg).Get(0).(plugin.Client)
}

type pluginClientMock struct {
	mock.Mock
}

func (m *pluginClientMock) Client() (goplugin.ClientProtocol, error) {
	called := m.Called()
	return called.Get(0).(goplugin.ClientProtocol), called.Error(1)
}

func (m *pluginClientMock) Kill() {
	m.Called()
}

type pluginClientProtocolMock struct {
	mock.Mock
}

func (m *pluginClientProtocolMock) Close() error {
	return m.Called().Error(0)
}

func (m *pluginClientProtocolMock) Dispense(arg string) (interface{}, error) {
	called := m.Called(arg)
	return called.Get(0), called.Error(1)
}

func (m *pluginClientProtocolMock) Ping() error {
	return m.Called().Error(0)
}

type grpcClientMock struct {
	drivertest.MockDriver
}

func (m *grpcClientMock) Stdout(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}

func (m *grpcClientMock) Stderr(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}

func setupPluginDriverTests(t *testing.T) (*grpcClientMock, func(*testing.T)) {
	execCommandMock := new(execCommandMock)
	plugin.ExecCommand = execCommandMock.Command
	newPluginClientMock := new(newPluginClientMock)
	pluginClientProtocolMock := new(pluginClientProtocolMock)
	pluginClientMock := new(pluginClientMock)
	grpcClientMock := new(grpcClientMock)
	grpcClientMock.On("Stdout", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	grpcClientMock.On("Stderr", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	plugin.NewPluginClient = newPluginClientMock.NewPlugin
	execCommandMock.On("Command", "fake-plugin-command").Return(
		func(string, ...string) *exec.Cmd {
			return new(exec.Cmd)
		},
	)
	newPluginClientMock.On("NewPlugin", mock.Anything).Return(pluginClientMock)
	pluginClientMock.On("Client").Return(pluginClientProtocolMock, nil)
	pluginClientMock.On("Kill").Return()
	pluginClientProtocolMock.On("Dispense", "driver_grpc").Return(grpcClientMock, nil)
	pluginClientProtocolMock.On("Ping").Return(nil)
	return grpcClientMock, func(t *testing.T) {
		plugin.ExecCommand = exec.Command
		plugin.NewPluginClient = plugin.DefaultPluginClient
	}
}

func TestPluginFieldResolve(t *testing.T) {
	grpcClientMock, teardown := setupPluginDriverTests(t)
	defer teardown(t)
	plug := plugin.NewPlugin(plugin.Config{
		Cmd: "fake-plugin-command",
	})
	defer plug.Close()
	data := []struct {
		in     driver.FieldResolveInput
		out    driver.FieldResolveOutput
		outErr error
	}{
		{
			in:  driver.FieldResolveInput{},
			out: driver.FieldResolveOutput{},
		},
		{
			in: driver.FieldResolveInput{},
			out: driver.FieldResolveOutput{
				Error: &driver.Error{
					Message: "",
				},
			},
			outErr: errors.New(""),
		},
	}
	for _, tt := range data {
		grpcClientMock.On("FieldResolve", tt.in).Return(tt.out, tt.outErr).Once()
		out := plug.FieldResolve(tt.in)
		assert.Equal(t, tt.out, out)
	}
	grpcClientMock.AssertNumberOfCalls(t, "FieldResolve", len(data))
}

func TestPluginInterfaceResolveType(t *testing.T) {
	grpcClientMock, teardown := setupPluginDriverTests(t)
	defer teardown(t)
	plug := plugin.NewPlugin(plugin.Config{
		Cmd: "fake-plugin-command",
	})
	defer plug.Close()
	data := []struct {
		in     driver.InterfaceResolveTypeInput
		out    driver.InterfaceResolveTypeOutput
		outErr error
	}{
		{
			in:  driver.InterfaceResolveTypeInput{},
			out: driver.InterfaceResolveTypeOutput{},
		},
		{
			in: driver.InterfaceResolveTypeInput{},
			out: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "",
				},
			},
			outErr: errors.New(""),
		},
	}
	for _, tt := range data {
		grpcClientMock.On("InterfaceResolveType", tt.in).Return(tt.out, tt.outErr).Once()
		out := plug.InterfaceResolveType(tt.in)
		assert.Equal(t, tt.out, out)
	}
	grpcClientMock.AssertNumberOfCalls(t, "InterfaceResolveType", len(data))
}

func TestPluginScalarParse(t *testing.T) {
	grpcClientMock, teardown := setupPluginDriverTests(t)
	defer teardown(t)
	plug := plugin.NewPlugin(plugin.Config{
		Cmd: "fake-plugin-command",
	})
	defer plug.Close()
	data := []struct {
		in     driver.ScalarParseInput
		out    driver.ScalarParseOutput
		outErr error
	}{
		{
			in:  driver.ScalarParseInput{},
			out: driver.ScalarParseOutput{},
		},
		{
			in: driver.ScalarParseInput{},
			out: driver.ScalarParseOutput{
				Error: &driver.Error{
					Message: "",
				},
			},
			outErr: errors.New(""),
		},
	}
	for _, tt := range data {
		grpcClientMock.On("ScalarParse", tt.in).Return(tt.out, tt.outErr).Once()
		out := plug.ScalarParse(tt.in)
		assert.Equal(t, tt.out, out)
	}
	grpcClientMock.AssertNumberOfCalls(t, "ScalarParse", len(data))
}

func TestPluginScalarSerialize(t *testing.T) {
	grpcClientMock, teardown := setupPluginDriverTests(t)
	defer teardown(t)
	plug := plugin.NewPlugin(plugin.Config{
		Cmd: "fake-plugin-command",
	})
	defer plug.Close()
	data := []struct {
		in     driver.ScalarSerializeInput
		out    driver.ScalarSerializeOutput
		outErr error
	}{
		{
			in:  driver.ScalarSerializeInput{},
			out: driver.ScalarSerializeOutput{},
		},
		{
			in: driver.ScalarSerializeInput{},
			out: driver.ScalarSerializeOutput{
				Error: &driver.Error{
					Message: "",
				},
			},
			outErr: errors.New(""),
		},
	}
	for _, tt := range data {
		grpcClientMock.On("ScalarSerialize", tt.in).Return(tt.out, tt.outErr).Once()
		out := plug.ScalarSerialize(tt.in)
		assert.Equal(t, tt.out, out)
	}
	grpcClientMock.AssertNumberOfCalls(t, "ScalarSerialize", len(data))
}

func TestPluginUnionResolveType(t *testing.T) {
	grpcClientMock, teardown := setupPluginDriverTests(t)
	defer teardown(t)
	plug := plugin.NewPlugin(plugin.Config{
		Cmd: "fake-plugin-command",
	})
	defer plug.Close()
	data := []struct {
		in     driver.UnionResolveTypeInput
		out    driver.UnionResolveTypeOutput
		outErr error
	}{
		{
			in:  driver.UnionResolveTypeInput{},
			out: driver.UnionResolveTypeOutput{},
		},
		{
			in: driver.UnionResolveTypeInput{},
			out: driver.UnionResolveTypeOutput{
				Error: &driver.Error{
					Message: "",
				},
			},
			outErr: errors.New(""),
		},
	}
	for _, tt := range data {
		grpcClientMock.On("UnionResolveType", tt.in).Return(tt.out, tt.outErr).Once()
		out := plug.UnionResolveType(tt.in)
		assert.Equal(t, tt.out, out)
	}
	grpcClientMock.AssertNumberOfCalls(t, "UnionResolveType", len(data))
}

func TestPluginSecrets(t *testing.T) {
	grpcClientMock, teardown := setupPluginDriverTests(t)
	defer teardown(t)
	grpcClientMock.On("FieldResolve", driver.FieldResolveInput{}).Return(
		driver.FieldResolveOutput{},
		nil,
	)
	plug := plugin.NewPlugin(plugin.Config{
		Cmd: "fake-plugin-command",
	})
	defer plug.Close()
	out := plug.SetSecrets(driver.SetSecretsInput{
		Secrets: driver.Secrets{
			"SECRET_VAR": "value",
		},
	})
	assert.Nil(t, out.Error)
	plug.FieldResolve(driver.FieldResolveInput{})
	out = plug.SetSecrets(driver.SetSecretsInput{
		Secrets: driver.Secrets{
			"SECRET_VAR": "value",
		},
	})
	assert.NotNil(t, out.Error)
}

func TestPluginStream(t *testing.T) {
	grpcClientMock, teardown := setupPluginDriverTests(t)
	defer teardown(t)
	plug := plugin.NewPlugin(plugin.Config{
		Cmd: "fake-plugin-command",
	})
	defer plug.Close()
	data := []struct {
		in     driver.StreamInput
		out    driver.StreamOutput
		outErr error
	}{
		{
			in:  driver.StreamInput{},
			out: driver.StreamOutput{},
		},
		{
			in: driver.StreamInput{},
			out: driver.StreamOutput{
				Error: &driver.Error{
					Message: "",
				},
			},
			outErr: errors.New(""),
		},
	}
	for _, tt := range data {
		grpcClientMock.On("Stream", tt.in).Return(tt.out, tt.outErr).Once()
		out := plug.Stream(tt.in)
		assert.Equal(t, tt.out, out)
	}
	grpcClientMock.AssertNumberOfCalls(t, "Stream", len(data))
}

type execCommandContextMock struct {
	mock.Mock
}

func (m *execCommandContextMock) CommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
	i := []interface{}{ctx, command}
	for _, arg := range args {
		i = append(i, arg)
	}
	called := m.Called(i...)
	f := called.Get(0).(func(ctx context.Context, command string, args ...string) *exec.Cmd)
	return f(ctx, command, args...)
}

func fakeBadExecCommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestLoadDriverPluginsCallsConfigHelperBad", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func fakeExecCommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestLoadDriverPluginsCallsConfigHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestLoadDriverPluginsCallsConfigHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	json.NewEncoder(os.Stdout).Encode([]driver.Config{
		driver.Config{
			Provider: "fake",
			Runtime:  "fake",
		},
	})
	os.Exit(0)
}

func TestLoadDriverPluginsCallsConfigHelperBad(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(1)
}
