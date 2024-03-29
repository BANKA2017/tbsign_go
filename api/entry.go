package _api

import (
	"github.com/labstack/echo/v4"
)

func Api(address string, variables ...any) {
	// api
	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(SetHeaders)

	// pre-check
	e.Use(PreCheck)

	// set variables
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
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
	e.PUT("/passport/update_pwd", UpdatePassword)
	e.GET("/passport/settings", GetSettings)
	e.PUT("/passport/settings", UpdateSettings)

	// tieba account
	e.GET("/account", GetTiebaAccountList)
	e.PUT("/account", AddTiebaAccount)
	e.DELETE("/account/:pid", RemoveTiebaAccount)
	e.GET("/account/status/:pid", CheckTiebaAccount)

	// tieba list
	e.GET("/list", GetTiebaList)
	e.PUT("/list", AddTieba)
	e.DELETE("/list/:pid/:fid", RemoveTieba)
	e.PUT("/list/ignore/:pid/:fid", IgnoreTieba)
	e.DELETE("/list/ignore/:pid/:fid", IgnoreTieba)
	e.POST("/list/refresh", RefreshTiebaList)
	e.POST("/list/clean", CleanTiebaList)

	// manage
	e.GET("/admin/settings", GetAdminSettings)
	e.PUT("/admin/settings", UpdateAdminSettings)
	e.GET("/admin/account", GetAccountList)

	// server status
	e.GET("/server/status", GetServerStatus)

	// plugins
	e.GET("/plugins/ver4_rank/list", GetVer4RankList)
	e.GET("/plugins/ver4_rank/settings", GetVer4RankSettings)
	e.PUT("/plugins/ver4_rank/settings", UpdateVer4RankSettings)

	e.GET("/plugins/ver4_ref/list", GetVer4RefList)
	e.POST("/plugins/ver4_ref/refresh", RefreshVer4RefTiebaList)

	// tools
	e.GET("/tool/tieba_uid2user_info/:tiebauid", GetUserByTiebaUID)

	e.Logger.Fatal(e.Start(address))
}
