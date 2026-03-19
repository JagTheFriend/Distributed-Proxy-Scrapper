package routes

import (
	"context"
	"encoding/json"
	"net/http"
	nodemanager "node-manager"
	"strings"
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
	Link     string `json:"link" validate:"required,url"`
	Method   string `json:"method" validate:"required"`
	ClientId string `json:"clientId,omitempty"`
}

func NewTriggerHanlder(e *echo.Group) *TriggerHandlerStruct {
	valkeyClient := nodemanager.GetValKeyClient()
	return &TriggerHandlerStruct{e: e, valkey: valkeyClient}
}

func (h *TriggerHandlerStruct) RegisterRoutes() {
	h.e.POST("/trigger/single", h.triggerSingle)
	h.e.POST("/trigger/multiple", h.triggerMultiple)
}

// Normalize & validate HTTP method
func normalizeMethod(method string) string {
	method = strings.ToUpper(method)
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
		return method
	default:
		return "GET"
	}
}

func (h *TriggerHandlerStruct) triggerSingle(c *echo.Context) error {
	var payload TriggerRequestPayload

	if err := c.Bind(&payload); err != nil {
		return err
	}

	ctx := context.Background()
	method := normalizeMethod(payload.Method)

	var clientId string

	// 1. Use provided clientId
	if payload.ClientId != "" {
		clientId = payload.ClientId
	} else {
		// 2. Pick available node
		key, err := nodemanager.GetAvailableClient(ctx)
		if err != nil {
			return echo.NewHTTPError(500, "No nodes available")
		}
		clientId = key[len("client:available:"):]
	}

	// mark busy
	err := nodemanager.SetClientBusy(ctx, clientId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to mark node as busy")
	}

	// get WS connection
	wsMutex.Lock()
	ws, ok := wsClients[clientId]
	wsMutex.Unlock()

	if !ok {
		return echo.NewHTTPError(500, "Node not connected")
	}

	jobId := uuid.NewString()

	req := map[string]interface{}{
		"type":  "JOB",
		"jobId": jobId,
		"payload": map[string]interface{}{
			"method": method,
			"url":    payload.Link,
		},
	}

	data, _ := json.Marshal(req)

	ch := make(chan ResultMessage)

	wsMutex.Lock()
	jobChannels[jobId] = ch
	wsMutex.Unlock()

	defer func() {
		wsMutex.Lock()
		delete(jobChannels, jobId)
		wsMutex.Unlock()
	}()

	if err := websocket.Message.Send(ws, string(data)); err != nil {
		return err
	}

	select {
	case res := <-ch:
		return c.JSON(http.StatusOK, res)
	case <-time.After(20 * time.Second):
		return echo.NewHTTPError(504, "Timeout")
	}
}

func (h *TriggerHandlerStruct) triggerMultiple(c *echo.Context) error {
	var payload TriggerRequestPayload
	if err := c.Bind(&payload); err != nil {
		return err
	}

	ctx := context.Background()
	method := normalizeMethod(payload.Method)

	// pick node
	key, err := nodemanager.GetAvailableClient(ctx)
	if err != nil {
		return echo.NewHTTPError(500, "No nodes available")
	}

	clientId := key[len("client:available:"):]

	// mark busy
	err = nodemanager.SetClientBusy(ctx, clientId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to mark node as busy")
	}

	// get WS
	wsMutex.Lock()
	ws, ok := wsClients[clientId]
	wsMutex.Unlock()

	if !ok {
		return echo.NewHTTPError(500, "Node not connected")
	}

	// create job
	jobId := uuid.NewString()

	req := map[string]interface{}{
		"type":  "JOB",
		"jobId": jobId,
		"payload": JobRequest{
			Method: method,
			URL:    payload.Link,
		},
	}

	data, _ := json.Marshal(req)

	// create channel
	ch := make(chan ResultMessage)
	wsMutex.Lock()
	jobChannels[jobId] = ch
	wsMutex.Unlock()

	defer func() {
		wsMutex.Lock()
		delete(jobChannels, jobId)
		wsMutex.Unlock()
	}()

	// send job
	if err := websocket.Message.Send(ws, string(data)); err != nil {
		return err
	}

	// wait response
	select {
	case res := <-ch:
		return c.JSON(http.StatusOK, res)
	case <-time.After(20 * time.Second):
		return echo.NewHTTPError(504, "Timeout")
	}
}
