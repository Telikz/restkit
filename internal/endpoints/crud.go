package endpoints

import (
	"context"
	"net/http"
	"strconv"

	rctx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/errors"
	"github.com/reststore/restkit/internal/middleware"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type PaginationParams struct {
	Limit  int32 `query:"limit"  default:"20"`
	Offset int32 `query:"offset" default:"0"`
	Page   int32 `query:"page"   default:"1"`
}

type ListParams struct {
	PaginationParams
	Sort  string `query:"sort"`
	Order string `query:"order"`
}

type SearchParams struct {
	PaginationParams
	Query string `query:"q"`
}

type GetRequest struct {
	ID int64 `path:"id"`
}

type DeleteRequest struct {
	ID int64 `path:"id"`
}

// ExtractParams extracts query and path parameters from a request type for OpenAPI docs.
func ExtractParams[Req any]() []Parameter {
	var params []Parameter

	queryParams := middleware.ExtractQueryParams[Req]()
	for _, p := range queryParams {
		params = append(params, Parameter{
			Name:     p.Name,
			Type:     p.Type,
			Required: p.Required,
			Location: ParamLocationQuery,
		})
	}

	pathParams := middleware.ExtractPathParams[Req]()
	for _, p := range pathParams {
		params = append(params, Parameter{
			Name:     p.Name,
			Type:     p.Type,
			Required: true,
			Location: ParamLocationPath,
		})
	}

	return params
}

func Get[Req any, Res any](
	path string,
	getFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		return getFn(ctx, req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "Get",
		Description: "Get a resource by ID",
		Handler:     handler,
		Bind:        middleware.QueryBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func GetWithQueries[Q any, Req any, Res any](
	path string,
	getFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		q, err := rctx.MustQueriesFromContext(ctx)
		if err != nil {
			return *new(Res), err
		}
		return getFn(ctx, q.(Q), req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "Get",
		Description: "Get a resource by ID",
		Handler:     handler,
		Bind:        middleware.QueryBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func List[Req any, Res any](
	path string,
	listFn func(ctx context.Context, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	handler := func(ctx context.Context, req Req) ([]Res, error) {
		list, err := listFn(ctx, req)
		if err != nil {
			return []Res{}, err
		}
		if list == nil {
			return []Res{}, nil
		}
		return list, nil
	}

	return &Endpoint[Req, []Res]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "List",
		Description: "List resources",
		Handler:     handler,
		Bind:        middleware.QueryBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func ListWithQueries[Q any, Req any, Res any](
	path string,
	listFn func(ctx context.Context, queries Q, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	handler := func(ctx context.Context, req Req) ([]Res, error) {
		q, err := rctx.MustQueriesFromContext(ctx)
		if err != nil {
			return []Res{}, err
		}

		list, err := listFn(ctx, q.(Q), req)
		if err != nil {
			return []Res{}, err
		}

		if list == nil {
			return []Res{}, nil
		}
		return list, nil
	}

	return &Endpoint[Req, []Res]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "List",
		Description: "List resources with optional filtering and pagination",
		Handler:     handler,
		Bind:        middleware.QueryBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func Create[Req any, Res any](
	path string,
	createFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		return createFn(ctx, req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodPost,
		Path:        path,
		Title:       "Create",
		Description: "Create a new resource",
		Handler:     handler,
		Bind:        middleware.JSONBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func CreateWithQueries[Q any, Req any, Res any](
	path string,
	createFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		q, err := rctx.MustQueriesFromContext(ctx)
		if err != nil {
			return *new(Res), err
		}
		return createFn(ctx, q.(Q), req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodPost,
		Path:        path,
		Title:       "Create",
		Description: "Create a new resource",
		Handler:     handler,
		Parameters:  ExtractParams[Req](),
	}
}

func Update[Req any, Res any](
	path string,
	updateFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		return updateFn(ctx, req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodPatch,
		Path:        path,
		Title:       "Update",
		Description: "Update a resource by ID",
		Handler:     handler,
		Bind:        middleware.MixedBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func UpdateWithQueries[Q any, Req any, Res any](
	path string,
	updateFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		q, err := rctx.MustQueriesFromContext(ctx)
		if err != nil {
			return *new(Res), err
		}
		return updateFn(ctx, q.(Q), req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodPatch,
		Path:        path,
		Title:       "Update",
		Description: "Update a resource by ID",
		Handler:     handler,
		Bind:        middleware.MixedBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func Delete[Req any](
	path string,
	deleteFn func(ctx context.Context, req Req) error,
) *Endpoint[Req, NoResponse] {
	handler := func(ctx context.Context, req Req) (NoResponse, error) {
		if err := deleteFn(ctx, req); err != nil {
			return NoResponse{}, err
		}
		return NoResponse{}, nil
	}

	return &Endpoint[Req, NoResponse]{
		Method:      http.MethodDelete,
		Path:        path,
		Title:       "Delete",
		Description: "Delete a resource by ID",
		Handler:     handler,
	}
}

func DeleteWithQueries[Q any, Req any](
	path string,
	deleteFn func(ctx context.Context, queries Q, req Req) error,
) *Endpoint[Req, NoResponse] {
	handler := func(ctx context.Context, req Req) (NoResponse, error) {
		q, err := rctx.MustQueriesFromContext(ctx)
		if err != nil {
			return NoResponse{}, err
		}
		return NoResponse{}, deleteFn(ctx, q.(Q), req)
	}

	return &Endpoint[Req, NoResponse]{
		Method:      http.MethodDelete,
		Path:        path,
		Title:       "Delete",
		Description: "Delete a resource by ID",
		Handler:     handler,
	}
}

func Search[Req any, Res any](
	path string,
	searchFn func(ctx context.Context, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	handler := func(ctx context.Context, req Req) ([]Res, error) {
		list, err := searchFn(ctx, req)
		if err != nil {
			return []Res{}, err
		}
		if list == nil {
			return []Res{}, nil
		}
		return list, nil
	}

	return &Endpoint[Req, []Res]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "Search",
		Description: "Search resources by query parameters",
		Handler:     handler,
		Bind:        middleware.QueryBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func SearchWithQueries[Q any, Req any, Res any](
	path string,
	searchFn func(ctx context.Context, queries Q, req Req) ([]Res, error),
) *Endpoint[Req, []Res] {
	handler := func(ctx context.Context, req Req) ([]Res, error) {
		q, err := rctx.MustQueriesFromContext(ctx)
		if err != nil {
			return []Res{}, err
		}
		list, err := searchFn(ctx, q.(Q), req)
		if err != nil {
			return []Res{}, err
		}
		if list == nil {
			return []Res{}, nil
		}
		return list, nil
	}

	return &Endpoint[Req, []Res]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "Search",
		Description: "Search resources by query parameters",
		Handler:     handler,
		Bind:        middleware.QueryBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func ParseID(idStr string) (int64, error) {
	return parseIDWithError(idStr)
}

func ParseIntID(idStr string) (int, error) {
	return parseIntIDWithError(idStr)
}

func parseIDWithError(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.ValidationFailed(
			http.StatusBadRequest, errors.ErrCodeBind,
			"Invalid id format", "id", "must be a valid integer",
		).ToAPIError()
	}
	return id, nil
}

func parseIntIDWithError(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.ValidationFailed(
			http.StatusBadRequest, errors.ErrCodeBind,
			"Invalid id format", "id", "must be a valid integer",
		).ToAPIError()
	}
	return id, nil
}
