module github.com/reststore/restkit/extra/grpc

go 1.26.1

require (
	github.com/reststore/restkit v0.0.0
	google.golang.org/grpc v1.67.0
)

require (
	golang.org/x/sys v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/reststore/restkit => ../..
