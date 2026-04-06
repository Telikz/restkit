module stdlib

go 1.26.1

require (
	github.com/reststore/restkit v0.0.0
	github.com/reststore/restkit/adapters/stdlib v0.0.0
	github.com/reststore/restkit/validators/playground v0.0.0
)

require (
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace (
	github.com/reststore/restkit => ../..
	github.com/reststore/restkit/adapters/stdlib => ../../adapters/stdlib
	github.com/reststore/restkit/validators/playground => ../../validators/playground
)
