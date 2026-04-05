module yaml

go 1.26.1

require (
	github.com/reststore/restkit v0.0.0
	github.com/reststore/restkit/serializers/yaml v0.0.0
)

require gopkg.in/yaml.v3 v3.0.1 // indirect

replace (
	github.com/reststore/restkit => ../..
	github.com/reststore/restkit/serializers/yaml => ../../serializers/yaml
)
