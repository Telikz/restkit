package endpoints

import (
	"context"
	"net/http"
	"strconv"

	rctx "github.com/reststore/restkit/internal/context"
	"github.com/reststore/restkit/internal/errors"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type PaginationParams struct {
	Limit  int32
	Offset int32
	Page   int32
}

type ListParams struct {
	PaginationParams
	Sort  string
	Order string
}

type SearchParams struct {
	PaginationParams
	Query string
}

func ListEndpoint[Q any, T any](
	path string,
	listFn func(ctx context.Context, queries Q, limit, offset int32) ([]T, error),
) *Endpoint[NoRequest, []T] {
	return &Endpoint[NoRequest, []T]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "List",
		Description: "List all resources with pagination",
		Parameters: []Parameter{
			{
				Name:        "limit",
				Type:        "integer",
				Description: "Number of items to return (default: 20)",
				Location:    ParamLocationQuery,
			},
			{
				Name:        "offset",
				Type:        "integer",
				Description: "Number of items to skip (default: 0)",
				Location:    ParamLocationQuery,
			},
		},
		Handler: func(ctx context.Context, _ NoRequest) ([]T, error) {
			limit := parseInt32Param(ctx, "limit", 20)
			offset := parseInt32Param(ctx, "offset", 0)
			list, err := listFn(ctx, rctx.MustQueriesFromContext(ctx).(Q), limit, offset)
			if err != nil {
				return []T{}, err
			}
			if list == nil {
				return []T{}, nil
			}
			return list, nil
		},
	}
}

func ListPaginatedEndpoint[Q any, T any](
	path string,
	listFn func(ctx context.Context, queries Q, params ListParams) ([]T, error),
) *Endpoint[NoRequest, []T] {
	return &Endpoint[NoRequest, []T]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "List",
		Description: "List resources with pagination and sorting",
		Parameters: []Parameter{
			{
				Name:        "limit",
				Type:        "integer",
				Description: "Number of items to return",
				Location:    ParamLocationQuery,
			},
			{
				Name:        "offset",
				Type:        "integer",
				Description: "Number of items to skip",
				Location:    ParamLocationQuery,
			},
			{
				Name:        "page",
				Type:        "integer",
				Description: "Page number",
				Location:    ParamLocationQuery,
			},
			{
				Name:        "sort",
				Type:        "string",
				Description: "Field to sort by",
				Location:    ParamLocationQuery,
			},
			{
				Name:        "order",
				Type:        "string",
				Description: "Sort order (asc or desc)",
				Location:    ParamLocationQuery,
			},
		},
		Handler: func(ctx context.Context, _ NoRequest) ([]T, error) {
			params := ListParams{
				PaginationParams: PaginationParams{
					Limit:  parseInt32Param(ctx, "limit", 20),
					Offset: parseInt32Param(ctx, "offset", 0),
					Page:   parseInt32Param(ctx, "page", 0),
				},
				Sort:  rctx.URLQueryParam(ctx, "sort"),
				Order: rctx.URLQueryParam(ctx, "order"),
			}
			list, err := listFn(ctx, rctx.MustQueriesFromContext(ctx).(Q), params)
			if err != nil {
				return []T{}, err
			}
			if list == nil {
				return []T{}, nil
			}
			return list, nil
		},
	}
}

func GetEndpoint[Q any, T any](
	path string,
	getFn func(ctx context.Context, queries Q, id int64) (T, error),
) *Endpoint[NoRequest, T] {
	return &Endpoint[NoRequest, T]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "Get",
		Description: "Get a resource by ID",
		Handler: func(ctx context.Context, _ NoRequest) (T, error) {
			var zero T
			id, err := ParseID(rctx.URLParam(ctx, "id"))
			if err != nil {
				return zero, err
			}
			return getFn(ctx, rctx.MustQueriesFromContext(ctx).(Q), id)
		},
	}
}

func CreateEndpoint[Q any, Req any, Res any](
	path string,
	createFn func(ctx context.Context, queries Q, req Req) (Res, error),
) *Endpoint[Req, Res] {
	return &Endpoint[Req, Res]{
		Method:      http.MethodPost,
		Path:        path,
		Title:       "Create",
		Description: "Create a new resource",
		Handler: func(ctx context.Context, req Req) (Res, error) {
			return createFn(ctx, rctx.MustQueriesFromContext(ctx).(Q), req)
		},
	}
}

func UpdateEndpoint[Q any, Req any](
	path string,
	updateFn func(ctx context.Context, queries Q, id int64, req Req) error,
) *Endpoint[Req, NoResponse] {
	return &Endpoint[Req, NoResponse]{
		Method:      http.MethodPatch,
		Path:        path,
		Title:       "Update",
		Description: "Update a resource by ID",
		Handler: func(ctx context.Context, req Req) (NoResponse, error) {
			id, err := ParseID(rctx.URLParam(ctx, "id"))
			if err != nil {
				return NoResponse{}, err
			}
			return NoResponse{}, updateFn(ctx, rctx.MustQueriesFromContext(ctx).(Q), id, req)
		},
	}
}

func DeleteEndpoint[Q any](
	path string,
	deleteFn func(ctx context.Context, queries Q, id int64) error,
) *Endpoint[NoRequest, MessageResponse] {
	return &Endpoint[NoRequest, MessageResponse]{
		Method:      http.MethodDelete,
		Path:        path,
		Title:       "Delete",
		Description: "Delete a resource by ID",
		Handler: func(ctx context.Context, _ NoRequest) (MessageResponse, error) {
			id, err := ParseID(rctx.URLParam(ctx, "id"))
			if err != nil {
				return MessageResponse{}, err
			}
			if err := deleteFn(ctx, rctx.MustQueriesFromContext(ctx).(Q), id); err != nil {
				return MessageResponse{}, err
			}
			return MessageResponse{Message: "deleted successfully"}, nil
		},
	}
}

func SearchEndpoint[Q any, T any](
	path string,
	searchFn func(ctx context.Context, queries Q) ([]T, error),
) *Endpoint[NoRequest, []T] {
	return &Endpoint[NoRequest, []T]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "Search",
		Description: "Search resources by query parameters",
		Parameters: []Parameter{
			{Name: "q", Type: "string", Description: "Search query", Location: ParamLocationQuery},
		},
		Handler: func(ctx context.Context, _ NoRequest) ([]T, error) {
			list, err := searchFn(ctx, rctx.MustQueriesFromContext(ctx).(Q))
			if err != nil {
				return []T{}, err
			}
			if list == nil {
				return []T{}, nil
			}
			return list, nil
		},
	}
}

func SearchPaginatedEndpoint[Q any, T any](
	path string,
	searchFn func(ctx context.Context, queries Q, params SearchParams) ([]T, error),
) *Endpoint[NoRequest, []T] {
	return &Endpoint[NoRequest, []T]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "Search",
		Description: "Search resources with pagination",
		Parameters: []Parameter{
			{Name: "q", Type: "string", Description: "Search query", Location: ParamLocationQuery},
			{
				Name:        "limit",
				Type:        "integer",
				Description: "Number of items to return",
				Location:    ParamLocationQuery,
			},
			{
				Name:        "offset",
				Type:        "integer",
				Description: "Number of items to skip",
				Location:    ParamLocationQuery,
			},
			{
				Name:        "page",
				Type:        "integer",
				Description: "Page number",
				Location:    ParamLocationQuery,
			},
		},
		Handler: func(ctx context.Context, _ NoRequest) ([]T, error) {
			params := SearchParams{
				PaginationParams: PaginationParams{
					Limit:  parseInt32Param(ctx, "limit", 20),
					Offset: parseInt32Param(ctx, "offset", 0),
					Page:   parseInt32Param(ctx, "page", 0),
				},
				Query: rctx.URLQueryParam(ctx, "q"),
			}
			list, err := searchFn(ctx, rctx.MustQueriesFromContext(ctx).(Q), params)
			if err != nil {
				return []T{}, err
			}
			if list == nil {
				return []T{}, nil
			}
			return list, nil
		},
	}
}

func ParseID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.ValidationFailed(
			http.StatusBadRequest, errors.ErrCodeBind,
			"Invalid id format", "id", "must be a valid integer",
		).ToAPIError()
	}
	return id, nil
}

func ParseIntID(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.ValidationFailed(
			http.StatusBadRequest, errors.ErrCodeBind,
			"Invalid id format", "id", "must be a valid integer",
		).ToAPIError()
	}
	return id, nil
}

func parseInt32Param(ctx context.Context, key string, defaultVal int32) int32 {
	s := rctx.URLQueryParam(ctx, key)
	if s == "" {
		return defaultVal
	}
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return defaultVal
	}
	return int32(i)
}
