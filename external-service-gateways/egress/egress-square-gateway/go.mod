module github.com/kofc7186/fundraiser-manager/external-service-gateways/egress/egress-square-gateway

go 1.21

replace github.com/kofc7186/fundraiser-manager => ../../../

require (
	cloud.google.com/go/pubsub v1.36.2
	github.com/GoogleCloudPlatform/functions-framework-go v1.8.1
	github.com/antihax/optional v1.0.0
	github.com/cloudevents/sdk-go/v2 v2.15.1
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/kofc7186/fundraiser-manager v0.0.0-00010101000000-000000000000
)

require (
	cloud.google.com/go v0.112.0 // indirect
	cloud.google.com/go/compute v1.24.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/functions v1.16.0 // indirect
	cloud.google.com/go/iam v1.1.6 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.1 // indirect
	github.com/googleapis/google-cloudevents-go v0.7.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.48.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.48.0 // indirect
	go.opentelemetry.io/otel v1.23.0 // indirect
	go.opentelemetry.io/otel/metric v1.23.0 // indirect
	go.opentelemetry.io/otel/trace v1.23.0 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/oauth2 v0.17.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/api v0.166.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240221002015-b0ce06bbee7c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/grpc v1.61.1 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)
