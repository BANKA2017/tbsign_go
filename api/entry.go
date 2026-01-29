package _api

import (
	"errors"
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

	// Why removed this middleware -> https://github.com/labstack/echo/issues/2211
	// TL;DR -> open embedded static dir in echo@v4 will cause incorrect redirection
	// e.Pre(middleware.RemoveTrailingSlash())

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var httpError *echo.HTTPError
		if ok := errors.As(err, &httpError); ok {
			if httpError.Code == http.StatusNotFound {
				_ = _function.EchoReject(c)
				return
			}
		}
		e.DefaultHTTPErrorHandler(err, c)
	}

	e.Use(ParsePath)

	apiPrefix := ""
	if share.EnableFrontend {
		apiPrefix = "/api"
	}

	// endpoints needn't check
	noCheckApi := e.Group(apiPrefix, SetHeaders)
	noCheckApi.POST("/passport/login", Login)
	noCheckApi.POST("/passport/signup", Signup)
	noCheckApi.POST("/passport/reset/password", ResetPassword, RateLimit(2, time.Second))
	noCheckApi.GET("/config/page/login", GetLoginPageConfig) // get site config before login

	api := e.Group(apiPrefix, SetHeaders, AuthCheck)

	// passport
	passport := api.Group("/passport")
	passport.GET("", GetAccountInfo)
	passport.POST("/logout", Logout)

	if share.EnableBackup {
		passport.POST("/export", ExportAccountData)
		passport.POST("/import", ImportAccountData)
	}

	passport.DELETE("/delete", DeleteAccount)
	passport.PUT("/update/info", UpdateAccountInfo)
	passport.PUT("/update/password", UpdatePassword)

	// tieba account
	tiebaAccount := api.Group("/account")
	tiebaAccount.GET("", GetTiebaAccountList)
	tiebaAccount.GET("/:pid", GetTiebaAccountItem)
	tiebaAccount.PATCH("", AddTiebaAccount)
	tiebaAccount.DELETE("/:pid", RemoveTiebaAccount)
	tiebaAccount.GET("/:pid/status", CheckTiebaAccount)
	tiebaAccount.GET("/qrcode", GetLoginQRCode)
	tiebaAccount.POST("/qrlogin", GetBDUSS)
	tiebaAccount.GET("/check/:pid/is_manager/:fname", CheckIsManager)

	// tieba list
	tiebaList := api.Group("/list")
	tiebaList.GET("", GetTiebaList)
	tiebaList.PATCH("", AddTieba)
	tiebaList.DELETE("", CleanTiebaList)
	tiebaList.POST("/sync", RefreshTiebaList, RateLimit(1, time.Second*10))
	tiebaList.DELETE("/:pid/:fid", RemoveTieba)
	tiebaList.PATCH("/:pid/:fid/ignore", IgnoreTieba)
	tiebaList.DELETE("/:pid/:fid/ignore", IgnoreTieba)
	tiebaList.POST("/:pid/:fid/reset", ResetTieba)

	// manage
	admin := api.Group("/admin", AdminCheck)
	admin.GET("/settings", GetAdminSettings)
	admin.POST("/settings", UpdateAdminSettings)
	admin.GET("/account", GetAccountsList)
	admin.DELETE("/account/:uid", AdminDeleteAccount)
	admin.PATCH("/account/:uid/modify", AdminModifyAccountInfo)
	admin.DELETE("/account/:uid/token", AdminDeleteAccountToken)
	admin.DELETE("/account/:uid/list", AdminDeleteTiebaAccountList)
	admin.POST("/account/:uid/list/reset", AdminResetTiebaList)
	admin.POST("/account/:uid/password/reset", AdminResetPassword)
	admin.POST("/account/:uid/plugin/:plugin_name/reset", AdminResetAccountPlugin)
	admin.POST("/plugin/:plugin_name/switch", PluginSwitch)
	admin.DELETE("/plugin/:plugin_name", PluginUninstall)
	admin.POST("/service/push/mail/test", SendTestMessage)

	/// server
	admin.GET("/server/status", GetServerStatus)
	admin.POST("/server/upgrade", UpgradeSystem)
	admin.POST("/server/shutdown", ShutdownSystem)

	if share.TestMode {
		hooks := admin.Group("/server/hooks")
		hooks.POST("/test/add-cron-time", HookAddCronTime)
	}

	// tools
	api.GET("/tools/userinfo/tieba_uid/:tiebauid", GetUserByTiebaUID)
	api.GET("/tools/userinfo/panel/:query_type/:user_value", GetUserByUsernameOrPortrait)
	api.GET("/tools/tieba/fname_to_fid/:fname", GetFidByFname)

	// notifications
	api.GET("/notifications", GetNotifications)

	// plugins
	plugin := api.Group("/plugins")
	plugin.GET("", GetPluginsList)
	plugin.Use(PluginPathPrecheck)
	for _, v := range _plugin.PluginList {
		for _, r := range v.(_plugin.PluginHooks).GetEndpoints() {
			plugin.Match([]string{r.Method}, "/"+v.(_plugin.PluginHooks).GetInfo().Name+"/"+r.Path, r.Function)
		}
	}

	// frontend
	if !share.EnableFrontend {
		e.Any("/favicon.ico", _function.EchoNoContent)
		e.Any("/robots.txt", echoRobots)
	} else {
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
