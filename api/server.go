package _api

import (
	"net/http"
	"os"
	"runtime"

	"github.com/labstack/echo/v4"
)

func GetServerStatus(c echo.Context) error {
	role := c.Get("role").(string)

	if role != "admin" {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid role", echoEmptyObject, "tbsign"))
	}

	hostname, _ := os.Hostname()

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"hostname":  hostname,
		"goroutine": runtime.NumGoroutine(),
		"version":   runtime.Version(),
		"variables": c.Get("variables"),
	}, "tbsign"))
}
