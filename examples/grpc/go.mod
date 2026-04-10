module github.com/reststore/restkit/examples/grpc

go 1.26.1

require (
	github.com/reststore/restkit v0.0.0
	github.com/reststore/restkit/extra/grpc v0.0.0
	google.golang.org/grpc v1.67.0
	google.golang.org/protobuf v1.34.2
)

require (
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
)

replace (
	github.com/reststore/restkit => ../..
	github.com/reststore/restkit/extra/grpc => ../../extra/grpc
)
