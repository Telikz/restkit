package websocket

import (
	"context"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/reststore/restkit/core"
)

// Upgrader is a function type that modifies the default WebSocket upgrader settings.
type Upgrader func(*ws.Upgrader)

// WebsocketHandler defines the signature for a user's websocket logic.
type WebsocketHandler[Req any] func(ctx context.Context, req Req, conn *ws.Conn) error

// defaultUpgrader is the default WebSocket upgrader with reasonable defaults.
var defaultUpgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// CheckOrigin allows all connections by default.
	CheckOrigin: func(r *http.Request) bool { return true },

	// Handle non-WebSocket upgrade requests gracefully
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(`{"error":"WebSocket upgrade required. Use ws:// or wss:// protocol"}`))
	},
}

// WebSocket creates an endpoint that upgrades the connection to a WebSocket.
func New[Req any](
	path string,
	wsFn WebsocketHandler[Req],
	upgraders ...Upgrader,
) *core.Endpoint[Req, core.NoResponse] {

	localUpgrader := defaultUpgrader
	for _, applyOpt := range upgraders {
		applyOpt(&localUpgrader)
	}

	e := &core.Endpoint[Req, core.NoResponse]{
		Method:  http.MethodGet,
		Path:    path,
		Scheme:  "ws",
		Title:   "WebSocket",
		Summary: "Websocket connection",
		Description: `WebSocket endpoint that upgrades the HTTP connection to a WebSocket.

Testing in swagger is not supported since Swagger UI does not support WebSocket connections. Use a WebSocket client like wscat or Postman to test this endpoint.`,
		Bind:       core.QueryBinder[Req](),
		Parameters: core.ExtractParams[Req](),
	}

	// Middleware to handle the WebSocket upgrade and connection management
	hijackMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := localUpgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			req, err := e.Bind(r)
			if err != nil {
				_ = conn.WriteMessage(ws.CloseMessage, []byte(err.Error()))
				return
			}

			if err := wsFn(r.Context(), req, conn); err != nil {
				_ = conn.WriteMessage(
					ws.CloseMessage,
					ws.FormatCloseMessage(ws.CloseInternalServerErr, err.Error()),
				)
			}
		})
	}

	// The handler is a no-op since the actual logic is handled in the middleware.
	e.Handler = func(ctx context.Context, req Req) (core.NoResponse, error) {
		return core.NoResponse{}, nil
	}

	// Attach the hijack middleware to handle WebSocket connections.
	return e.WithMiddleware(hijackMiddleware)
}
