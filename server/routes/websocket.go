package routes

import (
	"context"
	"encoding/json"
	"net/http"
	nodemanager "node-manager"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"golang.org/x/net/websocket"
)

type WSMessage struct {
	Type    string          `json:"type"`
	JobID   string          `json:"jobId,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type ResultMessage struct {
	Status   int    `json:"status"`
	Body     string `json:"body"`
	TimingMs int    `json:"timingMs"`
}

var (
	wsClients   = make(map[string]*websocket.Conn)
	wsMutex     sync.Mutex
	jobChannels = make(map[string]chan ResultMessage)
)

type WebSocketHandler struct {
	e      *echo.Group
	valkey *glide.Client
}

func NewWebSocketHandler(e *echo.Group) *WebSocketHandler {
	valkeyClient := nodemanager.GetValKeyClient()
	return &WebSocketHandler{e: e, valkey: valkeyClient}
}

func (h *WebSocketHandler) RegisterRoutes() {
	h.e.GET("/ws", h.websocketRoute)
}

func (h *WebSocketHandler) websocketRoute(c *echo.Context) error {
	clientId := c.QueryParam("clientId")
	clientType := c.QueryParam("clientType")

	if clientId == "" || clientType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing Query Params")
	}

	ctx := c.Request().Context()

	client := &nodemanager.Client{
		ClientId:    clientId,
		ClientType:  clientType,
		IP:          c.RealIP(),
		ConnectedAt: time.Now().Unix(),
	}

	if err := nodemanager.AddClient(ctx, client); err != nil {
		return err
	}

	defer nodemanager.RemoveClient(ctx, clientId)

	handler := websocket.Server{
		Handler: websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			// store connection
			wsMutex.Lock()
			wsClients[clientId] = ws
			wsMutex.Unlock()

			defer func() {
				wsMutex.Lock()
				delete(wsClients, clientId)
				wsMutex.Unlock()
			}()

			lastHeartbeat := time.Now()
			done := make(chan struct{})

			// heartbeat monitor
			go func() {
				for {
					select {
					case <-time.After(5 * time.Second):
						if time.Since(lastHeartbeat) > 15*time.Second {
							ws.Close()
							return
						}
					case <-done:
						return
					}
				}
			}()

			for {
				var msg WSMessage
				if err := websocket.JSON.Receive(ws, &msg); err != nil {
					close(done)
					return
				}

				switch msg.Type {

				case "PING":
					lastHeartbeat = time.Now()
					websocket.JSON.Send(ws, map[string]string{"type": "PONG"})

				case "RESULT":
					var res ResultMessage
					json.Unmarshal(msg.Payload, &res)

					wsMutex.Lock()
					ch, ok := jobChannels[msg.JobID]
					wsMutex.Unlock()

					if ok {
						ch <- res
					}

					// mark node idle again
					nodemanager.SetClientIdle(context.Background(), clientId)
				}
			}
		}),

		// THIS FIXES YOUR 403 ISSUE
		Handshake: func(config *websocket.Config, req *http.Request) error {
			return nil // allow all origins
		},
	}

	handler.ServeHTTP(c.Response(), c.Request())

	return nil
}
