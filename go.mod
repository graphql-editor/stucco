module github.com/graphql-editor/stucco

go 1.13

require (
	github.com/Dennor/gbtb v0.0.0-20191115154947-f9688184df1c
	github.com/blang/semver v3.5.1+incompatible
	github.com/golang/protobuf v1.3.2
	github.com/graphql-editor/azure-functions-golang-worker v0.1.0
	github.com/graphql-go/graphql v0.7.9-0.20191125031726-2e2b648ecbe4
	github.com/graphql-go/handler v0.2.3
	github.com/hashicorp/go-hclog v0.0.0-20180709165350-ff2cf002a8dd
	github.com/hashicorp/go-plugin v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v0.0.6
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa
	google.golang.org/grpc v1.26.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apiserver v0.17.3
	k8s.io/klog v1.0.0
)

replace github.com/blang/semver => github.com/blang/semver v1.1.1-0.20190414102917-ba2c2ddd8906

replace github.com/graphql-go/graphql => github.com/graphql-editor/graphql v0.7.10-0.20200602133915-4d19dee64a08
