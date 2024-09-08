package _function

import (
	"net/http"

	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
)

var EchoEmptyObject = make(map[string]any, 0)

func ApiTemplate[T any](code int, message string, data T, version string) _type.ApiTemplate {
	return _type.ApiTemplate{
		Code:    code,
		Message: message,
		Data:    data,
		Version: version,
	}
}

func EchoReject(c echo.Context) error {
	return c.JSON(http.StatusForbidden, ApiTemplate(403, "非法请求", EchoEmptyObject, "tbsign"))
}

func EchoNoContent(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
