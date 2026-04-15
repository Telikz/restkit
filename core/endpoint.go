package core

import (
	"context"

	ep "github.com/reststore/restkit/internal/endpoints"
)

// Endpoint is an alias for internal/endpoints.Endpoint. See restkit.Endpoint for details.
type Endpoint[Req any, Res any] = ep.Endpoint[Req, Res]

// NoRequest is an alias for internal/endpoints.NoRequest. See restkit.NoRequest for details.
type NoRequest = ep.NoRequest

// NoResponse is an alias for internal/endpoints.NoResponse. See restkit.NoResponse for details.
type NoResponse = ep.NoResponse

// GetRequest is an alias for internal/endpoints.GetRequest. See restkit.GetRequest for details.
type GetRequest = ep.GetRequest

// DeleteRequest is an alias for internal/endpoints.DeleteRequest. See restkit.DeleteRequest for details.
type DeleteRequest = ep.DeleteRequest

// MessageResponse is an alias for internal/endpoints.MessageResponse. See restkit.MessageResponse for details.
type MessageResponse = ep.MessageResponse

// Event is an alias for internal/endpoints.Event. See restkit.Event for details.
type Event[T any] = ep.Event[T]

// SearchRequest is an alias for internal/endpoints.SearchParams.
type SearchRequest = ep.SearchParams

// ListRequest is an alias for internal/endpoints.ListParams.
type ListRequest = ep.ListParams

// PaginationRequest is an alias for internal/endpoints.PaginationParams.
type PaginationRequest = ep.PaginationParams

// Endpoint constructors

// NewEndpoint creates a new endpoint with both request and response bodies.
func NewEndpoint[Req any, Res any]() *Endpoint[Req, Res] {
	return ep.NewEndpoint[Req, Res]()
}

// NewEndpointRes creates an endpoint that only returns a response body.
func NewEndpointRes[Res any]() *Endpoint[NoRequest, Res] {
	return ep.NewEndpointRes[Res]()
}

// NewEndpointReq creates an endpoint that only accepts a request body.
func NewEndpointReq[Req any]() *Endpoint[Req, NoResponse] {
	return ep.NewEndpointReq[Req]()
}

// Parameters

func ExtractParams[Req any]() []Parameter {
	return ep.ExtractParams[Req]()
}

// Parameter is an alias for internal/endpoints.Parameter. See restkit.Parameter for details.
type Parameter = ep.Parameter

// ParamLocation is an alias for internal/endpoints.ParamLocation.
type ParamLocation = ep.ParamLocation

const (
	ParamLocationPath  = ep.ParamLocationPath
	ParamLocationQuery = ep.ParamLocationQuery
)

// CRUD Endpoints

func List[Req any, Res any](
	path string,
	listFn func(ctx context.Context, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	return ep.List(path, listFn)
}

func Get[Req any, Res any](
	path string,
	getFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.Get(path, getFn)
}

func Post[Req any, Res any](
	path string,
	postFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.Post(path, postFn)
}

func Put[Req any, Res any](
	path string,
	putFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.Put(path, putFn)
}

func Patch[Req any, Res any](
	path string,
	patchFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.Patch(path, patchFn)
}

func Delete[Req any, Res any](
	path string,
	deleteFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.Delete(path, deleteFn)
}

// StreamEndpoint creates an endpoint for streaming resources using Server-Sent Events (SSE).
func Stream[Req any, Res any](
	path string,
	streamFn func(ctx context.Context, req Req) (<-chan ep.Event[Res], error),
) *Endpoint[Req, <-chan ep.Event[Res]] {
	return ep.Stream(path, streamFn)
}
