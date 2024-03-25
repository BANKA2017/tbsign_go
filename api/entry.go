package _api

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Api() {
	// api
	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Add("X-Powered-By", "tbsign")
			c.Response().Header().Add("Access-Control-Allow-Methods", "*")
			c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
			c.Response().Header().Add("Access-Control-Allow-Origin", "*")
			return next(c)
		}
	})

	// pre-check
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Println(c.Path(), c.QueryString())
			if PreCheckWhiteListExists(c.Path()) {
				//log.Println("echo: whitelist")
				return next(c)
			}
			uid := verifyAuthorization(c.Request().Header.Get("Authorization"))
			if uid == "0" {
				//log.Println("echo: invalid uid")
				return c.JSON(http.StatusOK, apiTemplate(403, "Invalid session", make(map[string]interface{}, 0), "tbsign"))
			}
			//log.Println("echo: next")
			c.Set("uid", uid)
			return next(c)
		}
	})

	e.Any("/*", echoReject)
	e.Any("/favicon.ico", echoFavicon)
	e.Any("/robots.txt", echoRobots)

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
	e.GET("/account/check", CheckTiebaAccount)

	// tieba list
	e.GET("/list", GetTiebaList)
	e.POST("/list/add", AddTieba)
	e.POST("/list/del", RemoveTieba)
	e.POST("/list/refresh", RefreshTiebaList)

	// manage
	e.GET("/admin/settings", GetAdminSettings)
	e.POST("/admin/settings", UpdateAdminSettings)
	e.GET("/admin/account", GetAccountList)

	// plugins
	e.GET("/plugins/ver4_rank/list", GetVer4RankList)
	e.GET("/plugins/ver4_rank/settings", GetVer4RankSettings)
	e.POST("/plugins/ver4_rank/settings", UpdateVer4RankSettings)

	// TODO tools

	e.Logger.Fatal(e.Start(":1323"))
}
