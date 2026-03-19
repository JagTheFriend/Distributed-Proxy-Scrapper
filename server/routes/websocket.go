package routes

import (
	"fmt"
	"net/http"
	nodemanager "node-manager"
	"sync"
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

			// Add client
			if err := nodemanager.AddClient(ctx, clientId); err != nil {
				c.Logger().Error("Failed to add client", "message", err.Error())
			}

			// Ensure removal once
			removed := false
			removeClient := func() {
				if !removed {
					removed = true
					if err := nodemanager.RemoveClient(ctx, clientId); err != nil {
						c.Logger().Error("Failed to remove client", "message", err.Error())
					}
				}
			}
			defer removeClient()

			heartbeatTimeout := 10 * time.Second
			lastHeartbeat := time.Now()

			// mutex to avoid race
			var mu sync.Mutex

			done := make(chan struct{})

			// Heartbeat monitor
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						mu.Lock()
						expired := time.Since(lastHeartbeat) > heartbeatTimeout
						mu.Unlock()

						if expired {
							c.Logger().Error("Heartbeat timeout, closing connection")
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
					c.Logger().Error("WS receive error", "message", err.Error())
					close(done)
					return
				}

				// Handle heartbeat
				if msg == "heartbeat" {
					mu.Lock()
					lastHeartbeat = time.Now()
					mu.Unlock()
					continue
				}

				fmt.Println("Received:", msg)
			}
		},
	}.ServeHTTP(c.Response(), c.Request())

	return nil
}
