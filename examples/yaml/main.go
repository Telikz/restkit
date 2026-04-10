package main

import (
	"context"
	"fmt"

	rk "github.com/reststore/restkit"
	rkyml "github.com/reststore/restkit/serializers/yaml"
)

type Config struct {
	Name    string   `json:"name"    yaml:"name" validate:"required, min=3, max=50"`
	Version string   `json:"version" yaml:"version"`
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Tags    []string `json:"tags"    yaml:"tags"`
}

func main() {
	api := rk.NewApi()
	api.WithSwaggerUI()
	api.WithTitle("YAML API")
	api.WithSerializer(rkyml.Serializer())
	api.WithDeserializer(rkyml.Deserializer())

	getConfigEndpoint := rk.Get("/config/{id}",
		func(_ context.Context, req rk.GetRequest) (Config, error) {
			return Config{
				Name:    fmt.Sprint(req.ID) + " - yamldemo",
				Version: "2.0.0",
				Enabled: true,
				Tags:    []string{"yaml", "config", "demo"},
			}, nil
		},
	)

	postConfigEndpoint := rk.Post("/config",
		func(_ context.Context, req Config) (Config, error) {
			req.Version = "processed-" + req.Version
			return req, nil
		},
	)

	api.AddEndpoint(getConfigEndpoint, postConfigEndpoint)

	fmt.Println("Server running on http://localhost:8080")
	if err := api.Serve(":8080"); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
