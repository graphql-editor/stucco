module github.com/graphql-editor/stucco

go 1.12

require (
	github.com/Dennor/gbtb v0.0.0-20191115154947-f9688184df1c
	github.com/blang/semver v3.5.0+incompatible
	github.com/golang/protobuf v1.3.2
	github.com/graphql-go/graphql v0.7.9-0.20191125031726-2e2b648ecbe4
	github.com/graphql-go/handler v0.2.3
	github.com/hashicorp/go-hclog v0.0.0-20180709165350-ff2cf002a8dd
	github.com/hashicorp/go-plugin v1.0.1
	github.com/pkg/errors v0.8.1
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	google.golang.org/grpc v1.23.1
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/apiserver v0.0.0-20191123100217-e01ab74ca9ea
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.1.0
)

replace github.com/blang/semver => github.com/lfaoro/semver v1.1.1-0.20190822180624-8f0c651cedf4
