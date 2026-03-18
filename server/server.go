package server

import (
	"common"
	"net/http"
	"server/routes"
	"time"

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
	e.Use(middleware.Recover())
	e.Use(middleware.ContextTimeout(time.Second * 5))

	groupedRoute := e.Group("/api/v1")
	websocketHandler := routes.NewWebSocketHandler(groupedRoute)
	websocketHandler.RegisterRoutes()

	groupedRoute.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("Failed to start server", "error", err)
	}
}
