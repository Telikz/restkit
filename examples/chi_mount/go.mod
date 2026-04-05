module chi_mount

go 1.26.1

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/reststore/restkit v0.0.0
	github.com/reststore/restkit/adapters/chi v0.0.0
)

replace (
	github.com/reststore/restkit => ../..
	github.com/reststore/restkit/adapters/chi => ../../adapters/chi
)
