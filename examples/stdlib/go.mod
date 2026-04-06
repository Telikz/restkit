module stdlib

go 1.26.1

require (
	github.com/reststore/restkit v0.0.0
	github.com/reststore/restkit/adapters/stdlib v0.0.0
)

replace (
	github.com/reststore/restkit => ../..
	github.com/reststore/restkit/adapters/stdlib => ../../adapters/stdlib
)
