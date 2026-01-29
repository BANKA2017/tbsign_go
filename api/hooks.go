package _api

import (
	"net/http"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v4"
)

func HookAddCronTime(c echo.Context) error {
	share.CrontabBypassTimes.Add(1)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{}, "tbsign"))
}
