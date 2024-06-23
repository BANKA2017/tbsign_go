package _api

import (
	"net/http"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
)

func GetNotifications(c echo.Context) error {
	notifications := _function.GetOption("ann")
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", notifications, "tbsign"))
}
