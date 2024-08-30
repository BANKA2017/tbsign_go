package _api

import (
	"io/fs"
	"net/http"

	"github.com/BANKA2017/tbsign_go/assets"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v4"
)

func Api(address string, args ...any) {
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
			for i := 0; i < len(args); i += 2 {
				_variable[args[i].(string)] = args[i+1]
			}

			// ext
			_variable["dbmode"] = share.DBMode
			_variable["testmode"] = share.TestMode

			c.Set("variables", _variable)
			return next(c)
		}
	})

	if !share.EnableFrontend {
		api.Any("/", echoReject)
		api.Any("/*", echoReject)
		api.Any("/favicon.ico", echoNoContent)
		api.Any("/robots.txt", echoRobots)
		api.OPTIONS("/*", echoNoContent)
	}

	api.GET("/passport", GetAccountInfo)
	api.POST("/passport/login", Login)
	api.POST("/passport/logout", Logout)
	api.POST("/passport/signup", Signup)
	//api.POST("/passport/export", ExportAccountData)
	api.DELETE("/passport/delete", DeleteAccount)
	api.PUT("/passport/update/info", UpdateAccountInfo)
	api.PUT("/passport/update/password", UpdatePassword)
	api.GET("/passport/settings", GetSettings)
	api.PUT("/passport/settings", UpdateSettings)
	api.POST("/passport/reset/password", ResetPassword)

	// tieba account
	api.GET("/account", GetTiebaAccountList)
	api.GET("/account/:pid", GetTiebaAccountItem)
	api.PATCH("/account", AddTiebaAccount)
	api.DELETE("/account/:pid", RemoveTiebaAccount)
	api.GET("/account/:pid/status", CheckTiebaAccount)
	api.GET("/account/qrcode", GetLoginQRCode)
	api.POST("/account/qrlogin", GetBDUSS)

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
	api.POST("/admin/plugin/:plugin_name/switch", PluginSwitch)
	api.POST("/admin/service/push/mail/test", SendTestMail)
	/// server
	api.GET("/admin/server/status", GetServerStatus)
	api.POST("/admin/server/upgrade", UpgradeSystem)
	api.POST("/admin/server/shutdown", ShutdownSystem)

	// plugins
	// ForumSupport
	api.GET("/plugins/forum_support/switch", PluginForumSupportGetSwitch)
	api.POST("/plugins/forum_support/switch", PluginForumSupportSwitch)
	api.GET("/plugins/forum_support/list", PluginForumSupportGetCharactersList)
	api.GET("/plugins/forum_support/settings", PluginForumSupportGetSettings)
	api.PUT("/plugins/forum_support/settings", PluginForumSupportUpdateSettings)

	// RefreshTiebaList
	api.GET("/plugins/refresh_tieba_list/list", PluginRefreshTiebaListGetAccountList)
	api.POST("/plugins/refresh_tieba_list/refresh", PluginRefreshTiebaListRefreshTiebaList)

	// LoopBan
	api.GET("/plugins/loop_ban/switch", PluginLoopBanGetSwitch)
	api.POST("/plugins/loop_ban/switch", PluginLoopBanSwitch)
	api.GET("/plugins/loop_ban/reason", PluginLoopBanGetReason)
	api.PUT("/plugins/loop_ban/reason", PluginLoopBanSetReason)
	api.GET("/plugins/loop_ban/list", PluginLoopBanGetList)
	api.PATCH("/plugins/loop_ban/list", PluginLoopBanAddAccounts)
	api.DELETE("/plugins/loop_ban/list/:id", PluginLoopBanDelAccount)
	api.POST("/plugins/loop_ban/list/empty", PluginLoopBanDelAllAccounts)
	api.GET("/plugins/loop_ban/check/:pid/is_manager/:fname", PluginLoopBanPreCheckIsManager)

	// GrowthTasks
	api.GET("/plugins/growth_tasks/settings", PluginGrowthTasksGetSettings)
	api.PUT("/plugins/growth_tasks/settings", PluginGrowthTasksSetSettings)
	api.GET("/plugins/growth_tasks/list", PluginGrowthTasksGetList)
	api.PATCH("/plugins/growth_tasks/list", PluginGrowthTasksAddAccount)
	api.DELETE("/plugins/growth_tasks/list/:id", PluginGrowthTasksDelAccount)
	api.POST("/plugins/growth_tasks/list/empty", PluginGrowthTasksDelAllAccounts)
	api.GET("/plugins/growth_tasks/status/:pid", PluginGrowthTasksGetTasksStatus)

	// Lottery
	api.GET("/plugins/knows_lottery/switch", PluginKnowsLotteryGetSwitch)
	api.POST("/plugins/knows_lottery/switch", PluginKnowsLotterySwitch)
	api.GET("/plugins/knows_lottery/log", PluginKnowsLotteryGetLogs)

	// tools
	api.GET("/tools/userinfo/tieba_uid/:tiebauid", GetUserByTiebaUID)
	api.GET("/tools/userinfo/panel/:query_type/:user_value", GetUserByUsernameOrPortrait)
	api.GET("/tools/tieba/fname_to_fid/:fname", GetFidByFname)

	// notifications
	api.GET("/notifications", GetNotifications)

	// others
	api.GET("/plugins", GetPluginsList)
	api.GET("/config/page/login", GetLoginPageConfig)

	// frontend
	if share.EnableFrontend {
		fe, _ := fs.Sub(assets.EmbeddedFrontent, "dist")
		e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(fe))))
	}

	e.Logger.Fatal(e.Start(address))
}
