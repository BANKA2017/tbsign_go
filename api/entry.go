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
	e.POST("/passport/signup", Signup)
	//e.POST("/passport/export", ExportAccountData)
	e.DELETE("/passport/delete", DeleteAccount)
	e.PUT("/passport/update_email", UpdateEmail)
	e.PUT("/passport/update_pwd", UpdatePassword)
	e.GET("/passport/settings", GetSettings)
	e.PUT("/passport/settings", UpdateSettings)
	e.POST("/passport/reset_password", ResetPassword)

	// tieba account
	e.GET("/account", GetTiebaAccountList)
	e.GET("/account/:pid", GetTiebaAccountItem)
	e.PATCH("/account", AddTiebaAccount)
	e.DELETE("/account/:pid", RemoveTiebaAccount)
	e.GET("/account/:pid/status", CheckTiebaAccount)

	// tieba list
	e.GET("/list", GetTiebaList)
	e.PATCH("/list", AddTieba)
	e.DELETE("/list", CleanTiebaList)
	e.POST("/list/refresh", RefreshTiebaList)
	e.DELETE("/list/:pid/:fid", RemoveTieba)
	e.PUT("/list/:pid/:fid/ignore", IgnoreTieba)
	e.DELETE("/list/:pid/:fid/ignore", IgnoreTieba)

	// manage
	e.GET("/admin/settings", GetAdminSettings)
	e.POST("/admin/settings", UpdateAdminSettings)
	e.GET("/admin/account", GetAccountsList)
	e.PATCH("/admin/account/modify/:uid", AdminModifyAccountInfo)
	e.DELETE("/admin/account/token/:uid", AdminDeleteAccountToken)
	e.DELETE("/admin/account/list/:uid", AdminDeleteTiebaAccountList)
	e.POST("/admin/plugin/:plugin_name/switch", PluginSwitch)
	e.GET("/admin/server/status", GetServerStatus)
	e.POST("/admin/service/push/mail/test", SendTestMail)

	// plugins
	// ForumSupport
	e.GET("/plugins/forum_support/switch", PluginForumSupportGetSwitch)
	e.POST("/plugins/forum_support/switch", PluginForumSupportSwitch)
	e.GET("/plugins/forum_support/list", PluginForumSupportGetCharactersList)
	e.GET("/plugins/forum_support/settings", PluginForumSupportGetSettings)
	e.PUT("/plugins/forum_support/settings", PluginForumSupportUpdateSettings)

	// RefreshTiebaList
	e.GET("/plugins/refresh_tieba_list/list", PluginRefreshTiebaListGetAccountList)
	e.POST("/plugins/refresh_tieba_list/refresh", PluginRefreshTiebaListRefreshTiebaList)

	// LoopBan
	e.GET("/plugins/loop_ban/switch", PluginLoopBanGetSwitch)
	e.POST("/plugins/loop_ban/switch", PluginLoopBanSwitch)
	e.GET("/plugins/loop_ban/reason", PluginLoopBanGetReason)
	e.PUT("/plugins/loop_ban/reason", PluginLoopBanSetReason)
	e.GET("/plugins/loop_ban/list", PluginLoopBanGetList)
	e.PATCH("/plugins/loop_ban/list", PluginLoopBanAddAccounts)
	e.DELETE("/plugins/loop_ban/list/:id", PluginLoopBanDelAccount)
	e.POST("/plugins/loop_ban/list/empty", PluginLoopBanDelAllAccounts)
	e.GET("/plugins/loop_ban/check/:pid/is_manager/:fname", PluginLoopBanPreCheckIsManager)

	// GrowthTasks
	e.GET("/plugins/growth_tasks/settings", PluginGrowthTasksGetSettings)
	e.PUT("/plugins/growth_tasks/settings", PluginGrowthTasksSetSettings)
	e.GET("/plugins/growth_tasks/list", PluginGrowthTasksGetList)
	e.PATCH("/plugins/growth_tasks/list", PluginGrowthTasksAddAccount)
	e.DELETE("/plugins/growth_tasks/list/:id", PluginGrowthTasksDelAccount)
	e.POST("/plugins/growth_tasks/list/empty", PluginGrowthTasksDelAllAccounts)
	e.GET("/plugins/growth_tasks/status/:pid", PluginGrowthTasksGetTasksStatus)

	// tools
	e.GET("/tools/userinfo/tieba_uid/:tiebauid", GetUserByTiebaUID)
	e.GET("/tools/userinfo/panel/:query_type/:user_value", GetUserByUsernameOrPortrait)
	e.GET("/tools/tieba/fname_to_fid/:fname", GetFidByFname)

	// notifications
	e.GET("/notifications", GetNotifications)

	// others
	e.GET("/plugins", GetPluginsList)
	e.GET("/config/page/login", GetLoginPageConfig)

	e.Logger.Fatal(e.Start(address))
}
