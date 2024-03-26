package _api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func Api(variables ...any) {
	// api
	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Add("X-Powered-By", "tbsign")
			c.Response().Header().Add("Access-Control-Allow-Methods", "*")
			c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
			c.Response().Header().Add("Access-Control-Allow-Origin", "*")
			c.Response().Header().Add("Access-Control-Allow-Headers", "Authorization")
			return next(c)
		}
	})

	// pre-check
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("start_date", time.Now().UnixNano())
			log.Println(c.Path(), c.QueryString())
			if PreCheckWhiteListExists(c.Path()) {
				//log.Println("echo: whitelist")
				return next(c)
			}
			uid, role := verifyAuthorization(c.Request().Header.Get("Authorization"))

			// login
			if uid == "0" {
				//log.Println("echo: invalid uid")
				return c.JSON(http.StatusOK, apiTemplate(403, "Invalid session", echoEmptyObject, "tbsign"))
			}

			// admin
			if strings.HasPrefix(c.Path(), "/admin/") && role != "admin" {
				return c.JSON(http.StatusOK, apiTemplate(403, "Invalid role", echoEmptyObject, "tbsign"))
			}
			//log.Println("echo: next")
			c.Set("uid", uid)
			c.Set("role", role)

			// set variables
			_variable := make(map[string]any)
			for i := 0; i < len(variables); i += 2 {
				_variable[variables[i].(string)] = variables[i+1]
			}
			c.Set("variables", _variable)
			return next(c)
		}
	})

	e.Any("/*", echoReject)
	e.Any("/favicon.ico", echoNoContent)
	e.Any("/robots.txt", echoRobots)
	e.OPTIONS("/*", echoNoContent)

	e.GET("/passport", GetAccountInfo)
	e.POST("/passport/login", Login)
	e.POST("/passport/logout", Logout)
	// e.POST("/passport/register", Register)
	// e.POST("/passport/delete", DeleteAccount)
	e.POST("/passport/update_pwd", UpdatePassword)
	e.GET("/passport/settings", GetSettings)
	e.POST("/passport/settings", UpdateSettings)

	// tieba account
	e.GET("/account", GetTiebaAccountList)
	e.POST("/account/add", AddTiebaAccount)
	e.POST("/account/del", RemoveTiebaAccount)
	e.GET("/account/check/:pid", CheckTiebaAccount)

	// tieba list
	e.GET("/list", GetTiebaList)
	e.POST("/list/add", AddTieba)
	e.POST("/list/del", RemoveTieba)
	e.POST("/list/refresh", RefreshTiebaList)
	e.POST("/list/clean", CleanTiebaList)

	// manage
	e.GET("/admin/settings", GetAdminSettings)
	e.POST("/admin/settings", UpdateAdminSettings)
	e.GET("/admin/account", GetAccountList)

	// server status
	e.GET("/server/status", GetServerStatus)

	// plugins
	e.GET("/plugins/ver4_rank/list", GetVer4RankList)
	e.GET("/plugins/ver4_rank/settings", GetVer4RankSettings)
	e.POST("/plugins/ver4_rank/settings", UpdateVer4RankSettings)

	e.GET("/plugins/ver4_ref/list", GetVer4RefList)
	e.POST("/plugins/ver4_ref/refresh", RefreshVer4RefTiebaList)

	// TODO tools

	e.Logger.Fatal(e.Start(":1323"))
}
