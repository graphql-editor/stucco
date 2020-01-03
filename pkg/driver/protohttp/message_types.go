package protohttp

type protobufMessageContentType string

const (
	protobufContentType                                            = "application/x-protobuf"
	fieldResolveRequestMessage          protobufMessageContentType = "FieldResolveRequest"
	fieldResolveResponseMessage         protobufMessageContentType = "FieldResolveResponse"
	interfaceResolveTypeRequestMessage  protobufMessageContentType = "InterfaceResolveTypeRequest"
	interfaceResolveTypeResponseMessage protobufMessageContentType = "InterfaceResolveTypeResponse"
	scalarParseRequestMessage           protobufMessageContentType = "ScalarParseRequest"
	scalarParseResponseMessage          protobufMessageContentType = "ScalarParseResponse"
	scalarSerializeRequestMessage       protobufMessageContentType = "ScalarSerializeRequest"
	scalarSerializeResponseMessage      protobufMessageContentType = "ScalarSerializeResponse"
	setSecretsRequestMessage            protobufMessageContentType = "SetSecretsRequest"
	setSecretsResponseMessage           protobufMessageContentType = "SetSecretsResponse"
	unionResolveTypeRequestMessage      protobufMessageContentType = "UnionResolveTypeRequest"
	unionResolveTypeResponseMessage     protobufMessageContentType = "UnionResolveTypeResponse"
)

func (p protobufMessageContentType) String() string {
	return protobufContentType + "; message=" + string(p)
}
