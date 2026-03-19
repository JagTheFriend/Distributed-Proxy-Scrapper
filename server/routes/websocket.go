package routes

import (
	"fmt"
	"net/http"
	nodemanager "node-manager"
	"time"

	"github.com/labstack/echo/v5"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"golang.org/x/net/websocket"
)

type WebSocketHandler struct {
	e      *echo.Group
	valkey *glide.Client
}

func NewWebSocketHandler(e *echo.Group) *WebSocketHandler {
	valkeyClient := nodemanager.GetValKeyClient()

	return &WebSocketHandler{
		e:      e,
		valkey: valkeyClient,
	}
}

func (h *WebSocketHandler) RegisterRoutes() {
	h.e.GET("/ws", h.websocketRoute)
}

func (h *WebSocketHandler) websocketRoute(c *echo.Context) error {
	clientId := c.Request().Header.Get("ClientId")
	if clientId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing ClientId")
	}

	websocket.Server{
		Handler: func(ws *websocket.Conn) {
			defer ws.Close()

			ctx := c.Request().Context()

			// Add client on connect
			client := &nodemanager.Client{
				ClientId: clientId,
			}
			if err := nodemanager.AddClient(ctx, client.ClientId); err != nil {
				c.Logger().Error("Failed to add client", "message", err.Error())
			}

			// Ensure client removal on disconnect
			defer func() {
				if err := nodemanager.RemoveClient(ctx, clientId); err != nil {
					c.Logger().Error("Failed to remove client", "message", err.Error())
				}
			}()

			heartbeatTimeout := 10 * time.Second
			lastHeartbeat := time.Now()

			done := make(chan struct{})

			// Heartbeat monitor
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						if time.Since(lastHeartbeat) > heartbeatTimeout {
							c.Logger().Error("Heartbeat timeout, closing connection")
							nodemanager.RemoveClient(ctx, clientId)
							ws.Close()
							return
						}
					case <-done:
						return
					}
				}
			}()

			// Message loop
			for {
				var msg string
				if err := websocket.Message.Receive(ws, &msg); err != nil {
					c.Logger().Error("WS receive error:", err)
					close(done)
					return
				}

				// Handle heartbeat
				if msg == "heartbeat" {
					lastHeartbeat = time.Now()
					continue
				}

				// Handle normal message
				fmt.Println("Received:", msg)
			}
		},
	}.ServeHTTP(c.Response(), c.Request())

	return nil
}
