package _api

import (
	"log"
	"net/http"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v4"
)

func SetHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !share.EnableFrontend {
			c.Response().Header().Add("Access-Control-Allow-Origin", "*")
		}
		c.Response().Header().Add("X-Powered-By", "TbSignGo->")
		c.Response().Header().Add("Access-Control-Allow-Methods", "*")
		c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
		c.Response().Header().Add("Access-Control-Allow-Headers", "Authorization")
		return next(c)
	}
}

func PreCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("start_date", _function.Now.UnixNano())
		method := c.Request().Method
		path := c.Path()
		log.Println(method, path, c.Request().URL.Path, c.QueryString())

		if PreCheckWhiteListExists(path) {
			return next(c)
		}

		if share.EnableFrontend && strings.HasPrefix(path, "/api/") {
			path = strings.TrimPrefix(path, "/api")
		}

		authorization := c.Request().Header.Get("Authorization")

		uid, role := verifyAuthorization(authorization)

		// login
		if uid == "0" {
			return c.JSON(http.StatusOK, _function.ApiTemplate(401, "无效 session", _function.EchoEmptyObject, "tbsign"))
		}

		// deleted
		if role == "deleted" {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号已删除", _function.EchoEmptyObject, "tbsign"))
		}

		// TODO banned
		if role == "banned" {
			if !(path == "/passport" || strings.HasPrefix(path, "/passport/") || path == "/notifications") {
				return c.JSON(http.StatusOK, _function.ApiTemplate(403, "受限账号", _function.EchoEmptyObject, "tbsign"))
			}
		}

		// admin
		if strings.HasPrefix(path, "/admin/") && role != "admin" {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效用户组", _function.EchoEmptyObject, "tbsign"))
		}

		c.Set("uid", uid)
		c.Set("role", role)

		return next(c)
	}
}

func PluginPathPrecheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Path()

		if share.EnableFrontend && strings.HasPrefix(path, "/api/") {
			path = strings.TrimPrefix(path, "/api/plugins/")
		} else {
			path = strings.TrimPrefix(path, "/plugins/")
		}

		pluginName := strings.SplitN(path, "/", 2)[0]

		_pluginInfo, ok := _plugin.PluginList[pluginName]

		if !ok || !_pluginInfo.(_plugin.PluginHooks).GetSwitch() {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "插件不可用", _function.EchoEmptyObject, "tbsign"))
		}

		return next(c)
	}
}
