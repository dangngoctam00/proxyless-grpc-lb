module github.com/asishrs/proxyless-grpc-lb/xds-server

go 1.21

require (
	github.com/asishrs/proxyless-grpc-lb/common v0.0.0-20200824005131-5718db7e19f5
	github.com/envoyproxy/go-control-plane v0.12.0
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	go.uber.org/zap v1.15.0
	google.golang.org/grpc v1.65.0
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

require (
	cel.dev/expr v0.15.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cncf/xds/go v0.0.0-20240423153145-555b57ec207b // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.2 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/term v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.60.1 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.18.0 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20200821003339-5e75c0163111 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.0
	k8s.io/client-go => k8s.io/client-go v0.18.0
)
