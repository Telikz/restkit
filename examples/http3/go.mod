module github.com/reststore/restkit/examples/http3

go 1.26.1

require (
	github.com/reststore/restkit v0.0.0
	github.com/reststore/restkit/extra/http3 v0.0.0
)

require (
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.59.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace (
	github.com/reststore/restkit => ../..
	github.com/reststore/restkit/extra/http3 => ../../extra/http3
)
