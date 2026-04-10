package endpoints

import (
	"context"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/reststore/restkit/internal/errors"
	mw "github.com/reststore/restkit/internal/middleware"
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

	queryParams := mw.ExtractQueryParams[Req]()
	for _, p := range queryParams {
		params = append(params, Parameter{
			Name:     p.Name,
			Type:     p.Type,
			Required: p.Required,
			Location: ParamLocationQuery,
		})
	}

	pathParams := mw.ExtractPathParams[Req]()
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
		Summary:     "Get a resource",
		Description: "Look up a resource",
		Handler:     handler,
		Bind:        mw.QueryBinder[Req](),
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
		Summary:     "List resources",
		Description: "Retrieve a list of resources",
		Handler:     handler,
		Bind:        mw.QueryBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func Post[Req any, Res any](
	path string,
	postFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		return postFn(ctx, req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodPost,
		Path:        path,
		Title:       "Create",
		Summary:     "Create a new resource",
		Description: "Create a new resource",
		Handler:     handler,
		Bind:        mw.JSONBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func Put[Req any, Res any](
	path string,
	updateFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		return updateFn(ctx, req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodPut,
		Title:       "Update",
		Summary:     "Update a resource by ID",
		Description: "Update a resource by ID",
		Handler:     handler,
		Bind:        mw.JSONBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func Patch[Req any, Res any](
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
		Summary:     "Update a resource by ID",
		Description: "Update a resource by ID",
		Handler:     handler,
		Bind:        mw.MixedBinder[Req](),
		Parameters:  ExtractParams[Req](),
	}
}

func Delete[Req any, Res any](
	path string,
	deleteFn func(ctx context.Context, req Req) (Res, error),
) *Endpoint[Req, Res] {
	handler := func(ctx context.Context, req Req) (Res, error) {
		return deleteFn(ctx, req)
	}

	return &Endpoint[Req, Res]{
		Method:      http.MethodDelete,
		Path:        path,
		Title:       "Delete",
		Summary:     "Delete a resource by ID",
		Description: "Delete a resource by ID",
		Handler:     handler,
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

func ParseUUID(idStr string) ([16]byte, error) {
	return parseUUIDWithError(idStr)
}

func parseUUIDWithError(idStr string) ([16]byte, error) {
	var uuid [16]byte

	s := strings.ReplaceAll(idStr, "-", "")
	if len(s) != 32 {
		return uuid, errors.ValidationFailed(
			http.StatusBadRequest, errors.ErrCodeBind,
			"Invalid uuid format", "id", "must be a valid uuid",
		).ToAPIError()
	}

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return uuid, errors.ValidationFailed(
			http.StatusBadRequest, errors.ErrCodeBind,
			"Invalid uuid format", "id", "must be a valid uuid",
		).ToAPIError()
	}

	copy(uuid[:], bytes)
	return uuid, nil
}
