package core

import (
	rc "github.com/reststore/restkit/internal/context"
)

// RouteContext is an alias for internal/context.RouteContext. See restkit.RouteContext for details.
type RouteContext = rc.RouteContext

var (
	WithQueries            = rc.WithQueries
	QueriesFromContext     = rc.QueriesFromContext
	MustQueriesFromContext = rc.MustQueriesFromContext
	URLParam               = rc.URLParam
	URLQueryParam          = rc.URLQueryParam
	RouteCtxFromContext    = rc.RouteCtxFromContext
	ExtractPathParams      = rc.ExtractPathParams
)
