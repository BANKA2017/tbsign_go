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
		return c.JSON(http.StatusOK, apiTemplate(500, "未知错误", echoEmptyObject, "tbsign"))
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", response, "tbsign"))
}

func GetUserByUsernameOrPortrait(c echo.Context) error {
	queryType := c.Param("query_type")
	userValue := c.Param("user_value")

	response, err := _function.GetUserInfoByUsernameOrPortrait(queryType, userValue)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, apiTemplate(500, "未知错误", echoEmptyObject, "tbsign"))
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", response, "tbsign"))
}

func GetFidByFname(c echo.Context) error {
	fname := c.Param("fname")

	fid := _function.GetFid(fname)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"fname": fname,
		"fid":   fid,
	}, "tbsign"))
}
