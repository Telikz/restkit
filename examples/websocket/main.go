package main

import (
	"context"
	"fmt"
	"log"

	ws "github.com/gorilla/websocket"
	rk "github.com/reststore/restkit"
	"github.com/reststore/restkit/extra/websocket"
)

type ChatRequest struct {
	RoomID string `path:"roomID"`
	User   string `query:"user"`
}

func main() {
	api := rk.NewApi()
	api.WithSwaggerUI()

	wsEndpoint := websocket.New("/api/chat/{roomID}",
		func(_ context.Context, req ChatRequest, conn *ws.Conn) error {
			user := req.User

			if user == "" {
				user = "Anonymous"
			}

			welcomeMsg := fmt.Sprintf("Welcome %s to room %s!", user, req.RoomID)
			if err := conn.WriteMessage(ws.TextMessage, []byte(welcomeMsg)); err != nil {
				return err
			}
			for {
				msgType, msg, err := conn.ReadMessage()
				if err != nil {
					log.Printf("Client disconnected: %v", err)
					return nil
				}

				log.Printf("[%s in %s] says: %s", user, req.RoomID, string(msg))

				reply := fmt.Sprintf("Server received: %s", string(msg))
				if err := conn.WriteMessage(msgType, []byte(reply)); err != nil {
					return err
				}
			}
		},
	)

	api.AddEndpoint(wsEndpoint)
	api.WithServer("ws://localhost:8080", "WebSocket server testing not supported", nil)
	api.WithServer("http://localhost:8080", "HTTP server for Swagger UI", nil)

	log.Println("WebSocket server running on http://localhost:8080")
	log.Println("Swagger UI available at http://localhost:8080/swagger")

	if err := api.Serve(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
