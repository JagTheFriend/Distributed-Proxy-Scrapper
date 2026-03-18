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
	ClientID := c.Request().Header.Get("ClientID")
	if ClientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing data")
	}

	websocket.Server{
		Handshake: func(cfg *websocket.Config, r *http.Request) error {
			return nil
		},
		Handler: func(ws *websocket.Conn) {
			defer ws.Close()

			heartbeatTimeout := 10 * time.Second
			lastHeartbeat := time.Now()

			// Monitor heartbeat
			done := make(chan struct{})

			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						if time.Since(lastHeartbeat) > heartbeatTimeout {
							c.Logger().Error("heartbeat timeout, closing connection")
							ws.Close()
							return
						}
					case <-done:
						return
					}
				}
			}()

			for {
				var msg string
				if err := websocket.Message.Receive(ws, &msg); err != nil {
					c.Logger().Error("WS receive error", "error", err)
					close(done)
					return
				}

				// Handle heartbeat
				if msg == "heartbeat" {
					lastHeartbeat = time.Now()
					continue
				}

				fmt.Println("Received:", msg)
			}
		},
	}.ServeHTTP(c.Response(), c.Request())

	return nil
}
