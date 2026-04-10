package grpc

import (
	"context"
	"net/http"

	rk "github.com/reststore/restkit"
	mw "github.com/reststore/restkit/internal/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPC creates a POST endpoint for gRPC with automatic error code mapping.
// Client is passed first, then the handler receives (ctx, client, request).
func GRPC[Client any, Req any, Res any](
	path string,
	client Client,
	handler func(ctx context.Context, c Client, req Req) (Res, error),
) *rk.Endpoint[Req, Res] {
	return &rk.Endpoint[Req, Res]{
		Method:      http.MethodPost,
		Path:        path,
		Title:       "gRPC",
		Summary:     "gRPC Gateway endpoint",
		Description: "gRPC Gateway endpoint with automatic error mapping",
		Handler: func(ctx context.Context, req Req) (Res, error) {
			return handler(ctx, client, req)
		},
		Bind:    mw.JSONBinder[Req](),
		Write:   mw.JSONWriter[Res](),
		OnError: grpcErrorHandler,
	}
}

// grpcErrorHandler maps gRPC status codes to HTTP status codes.
func grpcErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.NotFound:
			w.WriteHeader(http.StatusNotFound)
		case codes.InvalidArgument:
			w.WriteHeader(http.StatusBadRequest)
		case codes.Unauthenticated:
			w.WriteHeader(http.StatusUnauthorized)
		case codes.PermissionDenied:
			w.WriteHeader(http.StatusForbidden)
		case codes.AlreadyExists:
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(`{"error":"` + s.Message() + `"}`))
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error":"` + err.Error() + `"}`))
}
