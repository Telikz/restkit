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

// ListEndpoint creates an endpoint for listing resources.
func ListEndpoint[Q any, Req any, Res any](
	path string,
	listFn func(ctx context.Context, queries Q, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	return ep.ListWithQueries(path, listFn)
}

func Get[Req any, Res any](
	path string,
	getFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.Get(path, getFn)
}

// GetEndpoint creates an endpoint for getting a single resource.
func GetEndpoint[Q any, Req any, Res any](
	path string,
	getFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.GetWithQueries(path, getFn)
}

func Create[Req any, Res any](
	path string,
	createFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.Create(path, createFn)
}

// CreateEndpoint creates an endpoint for creating resources.
func CreateEndpoint[Q any, Req any, Res any](
	path string,
	createFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.CreateWithQueries(path, createFn)
}

func Update[Req any, Res any](
	path string,
	updateFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.Update(path, updateFn)
}

// UpdateEndpoint creates an endpoint for updating resources.
func UpdateEndpoint[Q any, Req any, Res any](
	path string,
	updateFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return ep.UpdateWithQueries(path, updateFn)
}

func Delete[Req any](
	path string,
	deleteFn func(ctx context.Context, req Req) error,
) *Endpoint[Req, NoResponse] {
	return ep.Delete(path, deleteFn)
}

// DeleteEndpoint creates an endpoint for deleting resources.
func DeleteEndpoint[Q any, Req any](
	path string,
	deleteFn func(ctx context.Context, queries Q, req Req) error,
) *Endpoint[Req, NoResponse] {
	return ep.DeleteWithQueries(path, deleteFn)
}

func Search[Req any, Res any](
	path string,
	searchFn func(ctx context.Context, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	return ep.Search(path, searchFn)
}

// SearchEndpoint creates an endpoint for searching resources.
func SearchEndpoint[Q any, Req any, Res any](
	path string,
	searchFn func(ctx context.Context, queries Q, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	return ep.SearchWithQueries(path, searchFn)
}
