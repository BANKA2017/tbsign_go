package _api

import (
	"log"
	"net/http"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
)

const authorizationPrefix = "bearer "

func SetHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Add("X-Powered-By", "TbSignGo->")
		c.Response().Header().Add("Access-Control-Allow-Methods", "*")
		c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
		c.Response().Header().Add("Access-Control-Allow-Origin", "*")
		c.Response().Header().Add("Access-Control-Allow-Headers", "Authorization")
		return next(c)
	}
}

func PreCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("start_date", _function.Now.UnixNano())
		method := c.Request().Method
		path := c.Path()
		log.Println(method, path, c.QueryString())

		if PreCheckWhiteListExists(path) {
			return next(c)
		}

		authorization := c.Request().Header.Get("Authorization")

		lengthOfAuthorizationPrefix := len(authorizationPrefix)

		if len(authorization) <= lengthOfAuthorizationPrefix || !strings.EqualFold(authorizationPrefix, authorization[0:lengthOfAuthorizationPrefix]) {
			return c.JSON(http.StatusOK, apiTemplate(401, "无效 session", echoEmptyObject, "tbsign"))
		}

		uid, role := verifyAuthorization(authorization[lengthOfAuthorizationPrefix:])

		// login
		if uid == "0" {
			return c.JSON(http.StatusOK, apiTemplate(401, "无效 session", echoEmptyObject, "tbsign"))
		}

		// deleted
		if role == "deleted" {
			return c.JSON(http.StatusOK, apiTemplate(403, "帐号已删除", echoEmptyObject, "tbsign"))
		}

		// TODO banned
		if role == "banned" {
			if !(path == "/passport" || strings.HasPrefix(path, "/passport/") || path == "/notifications") {
				return c.JSON(http.StatusOK, apiTemplate(403, "受限帐号", echoEmptyObject, "tbsign"))
			}
		}

		// admin
		if strings.HasPrefix(path, "/admin/") && role != "admin" {
			return c.JSON(http.StatusOK, apiTemplate(403, "无效用户组", echoEmptyObject, "tbsign"))
		}

		c.Set("uid", uid)
		c.Set("role", role)

		return next(c)
	}
}
