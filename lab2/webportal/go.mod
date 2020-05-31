module webportal

go 1.14

require (
	github.com/gorilla/mux v1.7.4
	github.com/newrelic/go-agent/v3 v3.5.0
	github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin v1.0.0
	github.com/newrelic/go-agent/v3/integrations/nrgorilla v1.1.0
	github.com/newrelic/go-agent/v3/integrations/nrgrpc v1.1.0
	github.com/newrelic/go-agent/v3/integrations/nrlogrus v1.0.0
	github.com/sirupsen/logrus v1.6.0
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.24.0 // indirect
	local.packages/sharedapp v0.0.0-00010101000000-000000000000
)

replace local.packages/sharedapp => ../sharedapp
