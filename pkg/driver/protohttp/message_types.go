package protohttp

import (
	"fmt"
	"strings"
)

type protobufMessageContentType string

const (
	protobufContentType                                              = "application/x-protobuf"
	authorizeRequestMessage               protobufMessageContentType = "AuthorizeRequest"
	authorizeResponseMessage              protobufMessageContentType = "AuthorizeResponse"
	fieldResolveRequestMessage            protobufMessageContentType = "FieldResolveRequest"
	fieldResolveResponseMessage           protobufMessageContentType = "FieldResolveResponse"
	interfaceResolveTypeRequestMessage    protobufMessageContentType = "InterfaceResolveTypeRequest"
	interfaceResolveTypeResponseMessage   protobufMessageContentType = "InterfaceResolveTypeResponse"
	scalarParseRequestMessage             protobufMessageContentType = "ScalarParseRequest"
	scalarParseResponseMessage            protobufMessageContentType = "ScalarParseResponse"
	scalarSerializeRequestMessage         protobufMessageContentType = "ScalarSerializeRequest"
	scalarSerializeResponseMessage        protobufMessageContentType = "ScalarSerializeResponse"
	setSecretsRequestMessage              protobufMessageContentType = "SetSecretsRequest"
	setSecretsResponseMessage             protobufMessageContentType = "SetSecretsResponse"
	unionResolveTypeRequestMessage        protobufMessageContentType = "UnionResolveTypeRequest"
	unionResolveTypeResponseMessage       protobufMessageContentType = "UnionResolveTypeResponse"
	subscriptionConnectionRequestMessage  protobufMessageContentType = "SubscriptionConnectionRequest"
	subscriptionConnectionResponseMessage protobufMessageContentType = "SubscriptionConnectionResponse"
)

func (p protobufMessageContentType) String() string {
	return protobufContentType + "; message=" + string(p)
}

func (p protobufMessageContentType) checkContentType(contentType string) error {
	messageType, err := getMessageType(contentType)
	if err == nil && !strings.EqualFold(string(p), messageType) {
		err = fmt.Errorf("cannot unmarshal %s to %s", messageType, string(p))
	}
	return err
}
