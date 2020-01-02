package grpc_test

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/stretchr/testify/mock"
	googlegrpc "google.golang.org/grpc"
)

type driverClientMock struct {
	mock.Mock
}

func (m *driverClientMock) FieldResolve(ctx context.Context, in *proto.FieldResolveRequest, opts ...googlegrpc.CallOption) (*proto.FieldResolveResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*proto.FieldResolveResponse), called.Error(1)
}

func (m *driverClientMock) InterfaceResolveType(ctx context.Context, in *proto.InterfaceResolveTypeRequest, opts ...googlegrpc.CallOption) (*proto.InterfaceResolveTypeResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*proto.InterfaceResolveTypeResponse), called.Error(1)
}

func (m *driverClientMock) ScalarParse(ctx context.Context, in *proto.ScalarParseRequest, opts ...googlegrpc.CallOption) (*proto.ScalarParseResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*proto.ScalarParseResponse), called.Error(1)
}

func (m *driverClientMock) ScalarSerialize(ctx context.Context, in *proto.ScalarSerializeRequest, opts ...googlegrpc.CallOption) (*proto.ScalarSerializeResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*proto.ScalarSerializeResponse), called.Error(1)
}

func (m *driverClientMock) SetSecrets(ctx context.Context, in *proto.SetSecretsRequest, opts ...googlegrpc.CallOption) (*proto.SetSecretsResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*proto.SetSecretsResponse), called.Error(1)
}

func (m *driverClientMock) UnionResolveType(ctx context.Context, in *proto.UnionResolveTypeRequest, opts ...googlegrpc.CallOption) (*proto.UnionResolveTypeResponse, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(*proto.UnionResolveTypeResponse), called.Error(1)
}
func (m *driverClientMock) Stream(ctx context.Context, in *proto.StreamRequest, opts ...googlegrpc.CallOption) (proto.Driver_StreamClient, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(proto.Driver_StreamClient), called.Error(1)
}

func (m *driverClientMock) Stdout(ctx context.Context, in *proto.ByteStreamRequest, opts ...googlegrpc.CallOption) (proto.Driver_StdoutClient, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(proto.Driver_StdoutClient), called.Error(1)
}

func (m *driverClientMock) Stderr(ctx context.Context, in *proto.ByteStreamRequest, opts ...googlegrpc.CallOption) (proto.Driver_StderrClient, error) {
	i := []interface{}{ctx, in}
	for _, opt := range opts {
		i = append(i, opt)
	}
	called := m.Called(i...)
	resp := called.Get(0)
	if resp == nil {
		return nil, called.Error(1)
	}
	return resp.(proto.Driver_StderrClient), called.Error(1)
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
