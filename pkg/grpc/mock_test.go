package grpc_test

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/driver"
	protoDriverService "github.com/graphql-editor/stucco_proto/go/driver_service"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
	"github.com/stretchr/testify/mock"
	googlegrpc "google.golang.org/grpc"
)

type driverClientMock struct {
	mock.Mock
}

func (m *driverClientMock) FieldResolve(ctx context.Context, in *protoMessages.FieldResolveRequest, opts ...googlegrpc.CallOption) (*protoMessages.FieldResolveResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*protoMessages.FieldResolveResponse), called.Error(1)
}

func (m *driverClientMock) InterfaceResolveType(ctx context.Context, in *protoMessages.InterfaceResolveTypeRequest, opts ...googlegrpc.CallOption) (*protoMessages.InterfaceResolveTypeResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*protoMessages.InterfaceResolveTypeResponse), called.Error(1)
}

func (m *driverClientMock) ScalarParse(ctx context.Context, in *protoMessages.ScalarParseRequest, opts ...googlegrpc.CallOption) (*protoMessages.ScalarParseResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*protoMessages.ScalarParseResponse), called.Error(1)
}

func (m *driverClientMock) ScalarSerialize(ctx context.Context, in *protoMessages.ScalarSerializeRequest, opts ...googlegrpc.CallOption) (*protoMessages.ScalarSerializeResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*protoMessages.ScalarSerializeResponse), called.Error(1)
}

func (m *driverClientMock) SetSecrets(ctx context.Context, in *protoMessages.SetSecretsRequest, opts ...googlegrpc.CallOption) (*protoMessages.SetSecretsResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*protoMessages.SetSecretsResponse), called.Error(1)
}

func (m *driverClientMock) UnionResolveType(ctx context.Context, in *protoMessages.UnionResolveTypeRequest, opts ...googlegrpc.CallOption) (*protoMessages.UnionResolveTypeResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*protoMessages.UnionResolveTypeResponse), called.Error(1)
}
func (m *driverClientMock) Stream(ctx context.Context, in *protoMessages.StreamRequest, opts ...googlegrpc.CallOption) (protoDriverService.Driver_StreamClient, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(protoDriverService.Driver_StreamClient), called.Error(1)
}

func (m *driverClientMock) Stdout(ctx context.Context, in *protoMessages.ByteStreamRequest, opts ...googlegrpc.CallOption) (protoDriverService.Driver_StdoutClient, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(protoDriverService.Driver_StdoutClient), called.Error(1)
}

func (m *driverClientMock) Stderr(ctx context.Context, in *protoMessages.ByteStreamRequest, opts ...googlegrpc.CallOption) (protoDriverService.Driver_StderrClient, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(protoDriverService.Driver_StderrClient), called.Error(1)
}

func (m *driverClientMock) SubscriptionConnection(ctx context.Context, in *protoMessages.SubscriptionConnectionRequest, opts ...googlegrpc.CallOption) (*protoMessages.SubscriptionConnectionResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*protoMessages.SubscriptionConnectionResponse), called.Error(1)
}

func (m *driverClientMock) SubscriptionListen(ctx context.Context, in *protoMessages.SubscriptionListenRequest, opts ...googlegrpc.CallOption) (protoDriverService.Driver_SubscriptionListenClient, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(protoDriverService.Driver_SubscriptionListenClient), called.Error(1)
}

type fieldResolveMock struct {
	mock.Mock
}

func (m *fieldResolveMock) Handle(input driver.FieldResolveInput) (interface{}, error) {
	called := m.Called(input)
	return called.Get(0), called.Error(1)
}

type interfaceResolveTypeMock struct {
	mock.Mock
}

func (m *interfaceResolveTypeMock) Handle(input driver.InterfaceResolveTypeInput) (string, error) {
	called := m.Called(input)
	return called.String(0), called.Error(1)
}

type setSecretsMock struct {
	mock.Mock
}

func (m *setSecretsMock) Handle(input driver.SetSecretsInput) error {
	return m.Called(input).Error(0)
}

type scalarParseMock struct {
	mock.Mock
}

func (m *scalarParseMock) Handle(input driver.ScalarParseInput) (interface{}, error) {
	called := m.Called(input)
	return called.Get(0), called.Error(1)
}

type scalarSerializeMock struct {
	mock.Mock
}

func (m *scalarSerializeMock) Handle(input driver.ScalarSerializeInput) (interface{}, error) {
	called := m.Called(input)
	return called.Get(0), called.Error(1)
}

type unionResolveTypeMock struct {
	mock.Mock
}

func (m *unionResolveTypeMock) Handle(input driver.UnionResolveTypeInput) (string, error) {
	called := m.Called(input)
	return called.String(0), called.Error(1)
}
