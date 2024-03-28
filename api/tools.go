package _api

import (
	"log"
	"net/http"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
)

func GetUserByTiebaUID(c echo.Context) error {
	tiebauid := c.Param("tiebauid")

	response, err := _function.GetUserInfoByTiebaUID(tiebauid)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, apiTemplate(500, "Unknown error", echoEmptyObject, "tbsign"))
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", response, "tbsign"))
}
