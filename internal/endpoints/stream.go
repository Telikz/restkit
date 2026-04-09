package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	mw "github.com/reststore/restkit/internal/middleware"
)

func Stream[Req any, Res any](
	path string,
	streamFn func(ctx context.Context, req Req) (<-chan Event[Res], error),
) *Endpoint[Req, <-chan Event[Res]] {

	handler := func(ctx context.Context, req Req) (<-chan Event[Res], error) {
		return streamFn(ctx, req)
	}

	return &Endpoint[Req, <-chan Event[Res]]{
		Method:      http.MethodGet,
		Path:        path,
		Title:       "Stream",
		Description: "Stream events using Server-Sent Events (SSE)",

		Handler:    handler,
		Bind:       mw.QueryBinder[Req](),
		Write:      StreamWriter[Res](),
		Parameters: ExtractParams[Req](),
	}
}

type Event[T any] struct {
	ID    string `json:"id"`
	Event string `json:"event,omitempty"`
	Retry int    `json:"retry,omitempty"`
	Data  T      `json:"data"`
}

func StreamWriter[T any]() func(w http.ResponseWriter, stream <-chan Event[T]) error {
	return func(w http.ResponseWriter, stream <-chan Event[T]) error {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported",
				http.StatusInternalServerError)
			return nil
		}

		for event := range stream {
			if event.ID != "" {
				if _, err := fmt.Fprintf(w, "id: %s\n", event.ID); err != nil {
					return err
				}
			}
			if event.Event != "" {
				if _, err := fmt.Fprintf(w, "event: %s\n", event.Event); err != nil {
					return err
				}
			}
			if event.Retry > 0 {
				if _, err := fmt.Fprintf(w, "retry: %d\n", event.Retry); err != nil {
					return err
				}
			}

			switch v := any(event.Data).(type) {
			case string:
				if _, err := fmt.Fprintf(w, "data: %s\n\n", v); err != nil {
					return err
				}
			case []byte:
				if _, err := fmt.Fprintf(w, "data: %s\n\n", strings.TrimSpace(string(v))); err != nil {
					return err
				}
			default:
				jsonData, err := json.Marshal(v)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
					return err
				}
			}
			fmt.Fprintf(w, "\n")
			flusher.Flush()
		}
		return nil
	}
}
