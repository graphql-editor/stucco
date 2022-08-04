module github.com/graphql-editor/stucco

go 1.16

require (
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/Dennor/gbtb v0.0.0-20200805082447-36c86fc7c9cb
	github.com/blang/semver/v4 v4.0.0
	github.com/bmatcuk/doublestar v1.3.1 // indirect
	github.com/buildkite/interpolate v0.0.0-20200526001904-07f35b4ae251 // indirect
	github.com/emicklei/go-restful v2.16.0+incompatible // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0
	github.com/graphql-editor/stucco_proto v0.7.21
	github.com/graphql-go/graphql v0.8.0
	github.com/graphql-go/handler v0.2.3
	github.com/hashicorp/go-hclog v1.2.1
	github.com/hashicorp/go-plugin v1.4.4
	github.com/hashicorp/yamux v0.0.0-20211028200310-0bc27b27de87 // indirect
	github.com/kennygrant/sanitize v1.2.4
	github.com/logrusorgru/aurora v0.0.0-20200102142835-e9ef32dff381
	github.com/mattn/go-colorable v0.1.12
	github.com/mattn/go-ieproxy v0.0.7 // indirect
	github.com/mattn/go-isatty v0.0.14
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0
	github.com/oklog/run v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/rs/cors v1.8.2
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.2
	golang.org/x/net v0.0.0-20220708220712-1185a9018129 // indirect
	golang.org/x/sys v0.0.0-20220712014510-0a85c31ab51e // indirect
	google.golang.org/genproto v0.0.0-20220714211235-042d03aeabc9 // indirect
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apiserver v0.24.3
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/utils v0.0.0-20220713171938-56c0de1e6f5e // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/graphql-go/graphql => github.com/graphql-editor/graphql v0.7.10-0.20220715103515-dd2af00bb70d

replace github.com/hashicorp/go-plugin => ../go-plugin
