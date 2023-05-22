module github.com/numaproj/numaflow-sinks/argoworkflow

go 1.18

require (
	github.com/numaproj/numaflow v0.8.0
	github.com/numaproj/numaflow-go v0.4.5
	github.com/numaproj/numaflow-sinks/shared v0.0.0-20230302175848-bf7b9cf08aab
	github.com/prometheus/client_golang v1.14.0
	github.com/stretchr/testify v1.8.0
	go.uber.org/zap v1.24.0
	k8s.io/apimachinery v0.26.3
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230323212658-478b75c54725 // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/utils v0.0.0-20230313181309-38a27ef9d749 // indirect
)

//replace github.com/numaproj/numaflow-sinks/shared => ../shared
