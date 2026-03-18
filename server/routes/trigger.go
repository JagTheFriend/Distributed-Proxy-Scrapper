package routes

import (
	"net/http"
	nodemanager "node-manager"

	"github.com/labstack/echo/v5"
	glide "github.com/valkey-io/valkey-glide/go/v2"
)

type TriggerRequestPayload struct {
	Link        string `json:"link" validate:"required,url"`
	ClientId    string `json:"clientId,omitempty" validate:"omitempty,uuid4"`
	NoOfClients int    `json:"noOfClients,omitempty" validate:"omitempty,gt=0"`
}

type TriggerHandlerStruct struct {
	e      *echo.Group
	valkey *glide.Client
}

func NewTriggerHanlder(e *echo.Group) *TriggerHandlerStruct {
	valkeyClient := nodemanager.GetValKeyClient()

	return &TriggerHandlerStruct{
		e:      e,
		valkey: valkeyClient,
	}
}

func (h *TriggerHandlerStruct) RegisterRoutes() {
	h.e.POST("/trigger/single", h.triggerSingle)
	h.e.POST("/trigger/multiple", h.triggerMultiple)
}

func (h *TriggerHandlerStruct) triggerSingle(c *echo.Context) error {
	var payload TriggerRequestPayload
	if err := c.Bind(payload); err != nil {
		return err
	}
	if err := c.Validate(payload); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, payload)
}

func (h *TriggerHandlerStruct) triggerMultiple(c *echo.Context) error {
	var payload TriggerRequestPayload
	if err := c.Bind(payload); err != nil {
		return err
	}
	if err := c.Validate(payload); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, payload)

}
