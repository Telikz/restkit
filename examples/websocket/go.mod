module websocket

go 1.26.1

require (
	github.com/gorilla/websocket v1.5.3
	github.com/reststore/restkit v0.0.0
	github.com/reststore/restkit/extra/websocket v0.0.0
)

replace github.com/reststore/restkit => ../..

replace github.com/reststore/restkit/extra/websocket => ../../extra/websocket
