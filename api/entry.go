package _api

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v4"
)

func Api(address string) {
	// api
	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(SetHeaders)

	apiPrefix := ""
	if share.EnableFrontend {
		apiPrefix = "/api"
	}

	api := e.Group(apiPrefix)
	// pre-check
	api.Use(PreCheck)

	// set variables
	api.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_variable := make(map[string]any)
			// ext
			_variable["dbmode"] = share.DBMode
			_variable["dbversion"] = share.DBVersion
			_variable["tlsdb"] = share.DBTLSOption != "false" && share.DBTLSOption != ""
			_variable["testmode"] = share.TestMode

			c.Set("variables", _variable)
			return next(c)
		}
	})

	if !share.EnableFrontend {
		api.Any("/", _function.EchoReject)
		api.Any("/*", _function.EchoReject)
		api.Any("/favicon.ico", _function.EchoNoContent)
		api.Any("/robots.txt", echoRobots)
		api.OPTIONS("/*", _function.EchoNoContent)
	}

	api.GET("/passport", GetAccountInfo)
	api.POST("/passport/login", Login)
	api.POST("/passport/logout", Logout)
	api.POST("/passport/signup", Signup)

	if share.EnableBackup {
		api.POST("/passport/export", ExportAccountData)
		api.POST("/passport/import", ImportAccountData)
	}

	api.DELETE("/passport/delete", DeleteAccount)
	api.PUT("/passport/update/info", UpdateAccountInfo)
	api.PUT("/passport/update/password", UpdatePassword)
	//api.GET("/passport/settings", GetSettings)
	//api.PUT("/passport/settings", UpdateSettings)
	api.POST("/passport/reset/password", ResetPassword, RateLimit(2, time.Second))

	// tieba account
	api.GET("/account", GetTiebaAccountList)
	api.GET("/account/:pid", GetTiebaAccountItem)
	api.PATCH("/account", AddTiebaAccount)
	api.DELETE("/account/:pid", RemoveTiebaAccount)
	api.GET("/account/:pid/status", CheckTiebaAccount)
	api.GET("/account/qrcode", GetLoginQRCode)
	api.POST("/account/qrlogin", GetBDUSS)
	api.GET("/account/check/:pid/is_manager/:fname", CheckIsManager)

	// tieba list
	api.POST("/list/sync", RefreshTiebaList)
	api.GET("/list", GetTiebaList)
	api.PATCH("/list", AddTieba)
	api.DELETE("/list", CleanTiebaList)
	api.DELETE("/list/:pid/:fid", RemoveTieba)
	api.PATCH("/list/:pid/:fid/ignore", IgnoreTieba)
	api.DELETE("/list/:pid/:fid/ignore", IgnoreTieba)
	api.POST("/list/:pid/:fid/reset", ResetTieba)

	// manage
	api.GET("/admin/settings", GetAdminSettings)
	api.POST("/admin/settings", UpdateAdminSettings)
	api.GET("/admin/account", GetAccountsList)
	api.PATCH("/admin/account/modify/:uid", AdminModifyAccountInfo)
	api.DELETE("/admin/account/token/:uid", AdminDeleteAccountToken)
	api.DELETE("/admin/account/list/:uid", AdminDeleteTiebaAccountList)
	api.POST("/admin/account/list/:uid/reset", AdminResetTiebaList)
	api.POST("/admin/account/password/:uid/reset", AdminResetPassword)
	api.POST("/admin/plugin/:plugin_name/switch", PluginSwitch)
	api.DELETE("/admin/plugin/:plugin_name", PluginUninstall)
	api.POST("/admin/service/push/mail/test", SendTestMessage)
	/// server
	api.GET("/admin/server/status", GetServerStatus)
	api.POST("/admin/server/upgrade", UpgradeSystem)
	api.POST("/admin/server/shutdown", ShutdownSystem)

	// tools
	api.GET("/tools/userinfo/tieba_uid/:tiebauid", GetUserByTiebaUID)
	api.GET("/tools/userinfo/panel/:query_type/:user_value", GetUserByUsernameOrPortrait)
	api.GET("/tools/tieba/fname_to_fid/:fname", GetFidByFname)

	// notifications
	api.GET("/notifications", GetNotifications)

	// others
	api.GET("/plugins", GetPluginsList)
	api.GET("/config/page/login", GetLoginPageConfig)

	// plugins
	plugin := api.Group("/plugins")
	plugin.Use(PluginPathPrecheck)
	for _, v := range _plugin.PluginList {
		// TDOO disable endpoint before install?
		for _, r := range v.(_plugin.PluginHooks).GetEndpoints() {
			plugin.Match([]string{r.Method}, "/"+v.(_plugin.PluginHooks).GetInfo().Name+"/"+r.Path, r.Function)
		}
	}
	// frontend
	if share.EnableFrontend {
		fe, _ := fs.Sub(assets.EmbeddedFrontent, "dist")
		e.GET("/icp.jsonp", func(c echo.Context) error {
			return c.JSONP(200, "__GetICP", struct {
				ICP string `json:"icp"`
			}{
				ICP: _function.GetOption("icp"),
			})
		})
		e.GET("/*", echo.WrapHandler(http.FileServer(&_function.StaticFSWrapper{
			FileSystem:   http.FS(fe),
			FixedModTime: share.BuildAtTime,
		})))
	}

	e.Logger.Fatal(e.Start(address))
}
