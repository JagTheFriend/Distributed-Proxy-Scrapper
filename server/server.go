package server

import (
	"common"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func StartServer() {
	port, err := common.GetEnv("PORT")
	if err != nil {
		panic("PORT not set")
	}

	e := echo.New()
	e.Use(middleware.RequestLogger())

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("Failed to start server", "error", err)
	}
}
