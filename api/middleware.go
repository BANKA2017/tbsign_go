package _api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func SetHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Add("X-Powered-By", "TbSignGo!")
		c.Response().Header().Add("Access-Control-Allow-Methods", "*")
		c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
		c.Response().Header().Add("Access-Control-Allow-Origin", "*")
		c.Response().Header().Add("Access-Control-Allow-Headers", "Authorization")
		return next(c)
	}
}

func PreCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("start_date", time.Now().UnixNano())
		log.Println(c.Request().Method, c.Path(), c.QueryString())

		if PreCheckWhiteListExists(c.Path()) {
			return next(c)
		}

		authorization := c.Request().Header.Get("Authorization")
		if len(authorization) < 6 || !strings.EqualFold("basic ", authorization[0:6]) {
			return c.JSON(http.StatusOK, apiTemplate(401, "无效 session", echoEmptyObject, "tbsign"))
		}

		uid, role := verifyAuthorization(authorization[6:])

		// login
		if uid == "0" {
			return c.JSON(http.StatusOK, apiTemplate(401, "无效 session", echoEmptyObject, "tbsign"))
		}

		// banned
		if role == "banned" {
			return c.JSON(http.StatusOK, apiTemplate(403, "帐号已封禁", echoEmptyObject, "tbsign"))
		} else if role == "deleted" {
			return c.JSON(http.StatusOK, apiTemplate(403, "帐号已删除", echoEmptyObject, "tbsign"))
		}

		// admin
		if strings.HasPrefix(c.Path(), "/admin/") && role != "admin" {
			return c.JSON(http.StatusOK, apiTemplate(403, "无效用户组", echoEmptyObject, "tbsign"))
		}

		c.Set("uid", uid)
		c.Set("role", role)

		return next(c)
	}
}
