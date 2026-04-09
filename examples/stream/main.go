package main

import (
	"context"
	"fmt"
	"time"

	rk "github.com/reststore/restkit"
)

// streamRequest defines the request parameters for the streaming endpoint.
type streamRequest struct {
	ID    string `path:"id"`     // Path parameter for the stream ID
	Topic string `query:"topic"` // Query parameter for filtering the stream by topic
}

// streamResponse defines the structure of the data sent in each event.
type streamResponse struct {
	To      string `json:"to"`      // The ID of the stream recipient taken from the path parameter
	Message string `json:"message"` // The message content
}

func main() {
	api := rk.NewApi()
	api.WithVersion("1.0.0")
	api.WithTitle("Streaming API")
	api.WithDescription("Example of a streaming endpoint using Server-Sent Events (SSE)")
	api.WithSwaggerUI()

	// streamEndpoint defines a streaming endpoint at /stream/{id}
	// that sends events based on the streamRequest parameters.
	streamEndpoint := rk.Stream("/stream/{id}",
		func(_ context.Context, req streamRequest) (<-chan rk.Event[streamResponse], error) {
			stream := make(chan rk.Event[streamResponse]) // stream channel for sending events to the client

			// go routine to simulate sending events every second for 5 seconds
			go func() {
				defer close(stream) // Close the stream channel when done
				for i := range 5 {
					// Send an event to the stream channel with streamResponse data
					stream <- rk.Event[streamResponse]{
						ID:    fmt.Sprintf("%d", i),
						Event: "message",
						Data: streamResponse{
							To:      req.ID,
							Message: fmt.Sprintf("Message %d for topic %s", i, req.Topic)},
					}
					time.Sleep(1 * time.Second) // Simulate delay between events
				}
			}()

			return stream, nil // Return the stream channel to the endpoint handler
		})

	api.AddEndpoint(streamEndpoint)

	fmt.Println("Starting Server on http://localhost:8080...")
	fmt.Println("Swagger starting on http://localhost:8080/swagger")

	if err := api.Serve(":8080"); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
