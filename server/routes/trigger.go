package routes

import (
	"context"
	"encoding/json"
	"net/http"
	nodemanager "node-manager"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"golang.org/x/net/websocket"
)

type JobRequest struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

type TriggerHandlerStruct struct {
	e      *echo.Group
	valkey *glide.Client
}

type TriggerRequestPayload struct {
	Link        string `json:"link" validate:"required,url"`
	ClientId    string `json:"clientId,omitempty" validate:"omitempty"`
	NoOfClients int    `json:"noOfClients,omitempty" validate:"omitempty,gt=0"`
}

func NewTriggerHanlder(e *echo.Group) *TriggerHandlerStruct {
	valkeyClient := nodemanager.GetValKeyClient()
	return &TriggerHandlerStruct{e: e, valkey: valkeyClient}
}

func (h *TriggerHandlerStruct) RegisterRoutes() {
	h.e.POST("/trigger", h.triggerMultiple)
}

func (h *TriggerHandlerStruct) triggerMultiple(c *echo.Context) error {
	var payload TriggerRequestPayload
	if err := c.Bind(&payload); err != nil {
		return err
	}

	ctx := context.Background()

	// 1. pick node
	key, err := nodemanager.GetAvailableClient(ctx)
	if err != nil {
		return echo.NewHTTPError(500, "No nodes available")
	}

	clientId := key[len("client:available:"):]

	// 2. mark busy
	nodemanager.SetClientBusy(ctx, clientId)

	// 3. get WS
	wsMutex.Lock()
	ws, ok := wsClients[clientId]
	wsMutex.Unlock()

	if !ok {
		return echo.NewHTTPError(500, "Node not connected")
	}

	// 4. create job
	jobId := uuid.NewString()

	req := map[string]interface{}{
		"type":  "JOB",
		"jobId": jobId,
		"payload": JobRequest{
			Method: "GET",
			URL:    payload.Link,
		},
	}

	data, _ := json.Marshal(req)

	// 5. create channel
	ch := make(chan ResultMessage)
	wsMutex.Lock()
	jobChannels[jobId] = ch
	wsMutex.Unlock()

	defer func() {
		wsMutex.Lock()
		delete(jobChannels, jobId)
		wsMutex.Unlock()
	}()

	// 6. send job
	if err := websocket.Message.Send(ws, string(data)); err != nil {
		return err
	}

	// 7. wait response
	select {
	case res := <-ch:
		return c.JSON(http.StatusOK, res)

	case <-time.After(20 * time.Second):
		return echo.NewHTTPError(504, "Timeout")
	}
}
