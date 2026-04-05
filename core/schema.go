package core

import (
	docs "github.com/reststore/restkit/internal/docs"
	sc "github.com/reststore/restkit/internal/schema"
)

// GenerateOpenAPIFile generates OpenAPI spec file at specified location.
var GenerateOpenAPIFile = docs.CreateOpenAPIFile

// MountedRoute is an alias for internal/schema.MountedRoute. See restkit.MountedRoute for details.
type MountedRoute = sc.MountedRoute

// ParamInfo is an alias for internal/schema.ParamInfo. See restkit.ParamInfo for details.
type ParamInfo = sc.ParamInfo

// RouteInfo is an alias for internal/schema.RouteInfo. See restkit.RouteInfo for details.
type RouteInfo = sc.RouteInfo

// RouteMeta is an alias for internal/schema.RouteMeta. See restkit.RouteMeta for details.
type RouteMeta = sc.RouteMeta

// SchemaFrom generates a JSON Schema from a Go type using reflection.
func SchemaFrom[T any]() map[string]any {
	return sc.SchemaFrom[T]()
}
