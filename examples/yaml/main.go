package main

import (
	"context"
	"fmt"
	"net/http"

	rest "github.com/reststore/restkit"
	yaml "github.com/reststore/restkit/serializers/yaml"
)

type Config struct {
	Name    string   `json:"name"    yaml:"name"`
	Version string   `json:"version" yaml:"version"`
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Tags    []string `json:"tags"    yaml:"tags"`
}

func main() {
	api := rest.NewApi()
	api.WithTitle("YAML API")
	api.WithSwaggerUI("/swagger")
	api.WithSerializer(yaml.Serializer())
	api.WithDeserializer(yaml.Deserializer())

	// Converts path param to yaml and and return config as response
	api.AddEndpoint(rest.NewEndpoint[rest.GetRequest, Config]().
		WithPath("/config/{id}").
		WithMethod(http.MethodGet).
		WithHandler(func(ctx context.Context, req rest.GetRequest) (Config, error) {
			return Config{
				Name:    fmt.Sprint(req.ID) + " - yamldemo",
				Version: "2.0.0",
				Enabled: true,
				Tags:    []string{"yaml", "config", "demo"},
			}, nil
		}))

	// Converts Json body to yaml and returns it as response
	api.AddEndpoint(rest.NewEndpoint[Config, Config]().
		WithPath("/config").
		WithMethod(http.MethodPost).
		WithHandler(func(ctx context.Context, req Config) (Config, error) {
			req.Version = "processed-" + req.Version
			return req, nil
		}))

	api.Serve(":8080")
}
