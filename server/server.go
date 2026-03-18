package server

import (
	"common"
	"context"
	"log/slog"
	"net/http"
	"os"
	"server/routes"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.ErrBadRequest.Wrap(err)
	}
	return nil
}

func StartServer() {
	port, err := common.GetEnv("PORT")
	if err != nil {
		panic("PORT not set")
	}

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.ContextTimeout(time.Second * 5))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	e.Validator = &CustomValidator{validator: validator.New()}

	groupedRoute := e.Group("/api/v1")

	websocketHandler := routes.NewWebSocketHandler(groupedRoute)
	websocketHandler.RegisterRoutes()

	triggerHandler := routes.NewTriggerHanlder(groupedRoute)
	triggerHandler.RegisterRoutes()

	groupedRoute.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("Failed to start server", "error", err)
	}
}
