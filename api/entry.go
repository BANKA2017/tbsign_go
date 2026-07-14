package _api

import (
	"context"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Api(ctx context.Context, network, address string) {
	// api
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogMethod:    true,
		LogRoutePath: true,
		LogURI:       true,
		HandleError:  true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			slogHttpGroup := slog.Group("http",
				slog.Int("status", v.Status),
				slog.String("method", v.Method),
				slog.String("path", v.RoutePath),
				slog.String("uri", v.URI),
				slog.String("query", c.QueryString()),
			)

			if v.Error != nil {
				slog.Error("echo.error",
					slogHttpGroup,
					slog.String("error", v.Error.Error()),
				)
			} else {
				// if v.Status == http.StatusOK && v.RoutePath == "/*" {
				// 	return nil
				// }
				slog.Debug("echo.request", slogHttpGroup)
			}
			return nil
		},
	}))

	// Why removed this middleware -> https://github.com/labstack/echo/issues/2211
	// TL;DR -> open embedded static dir in echo@v4 will cause incorrect redirection
	// e.Pre(middleware.RemoveTrailingSlash())

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
	api.Any("/*", _function.EchoReject)

	// passport
	passport := api.Group("/passport")
	passport.GET("", GetAccountInfo)
	passport.POST("/logout", Logout)

	if share.EnableBackup {
		passport.POST("/export", ExportAccountData, RateLimit(1, time.Second*10))
		passport.POST("/import", ImportAccountData, RateLimit(1, time.Second*10), middleware.BodyLimit(bodyLimit50M))
	}

	passport.DELETE("/delete", DeleteAccount)
	passport.PUT("/update/info", UpdateAccountInfo)
	passport.PUT("/update/password", UpdatePassword)
	passport.POST("/plugin/:plugin_name/reset", ResetAccountPlugin)
	passport.POST("/plugin/:plugin_name/reset/:pid", ResetAccountPlugin)
	passport.POST("/plugin/:plugin_name/reset/:pid/:tid", ResetAccountPlugin)

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
	tiebaList.GET("/:pid", GetTiebaList)
	tiebaList.PATCH("", AddTieba)
	tiebaList.DELETE("", CleanTiebaList)
	tiebaList.DELETE("/:pid", CleanTiebaList)
	tiebaList.POST("/sync", RefreshTiebaList, RateLimit(1, time.Second*10))
	tiebaList.DELETE("/:pid/:fid", RemoveTieba)
	tiebaList.PATCH("/:pid/:fid/ignore", IgnoreTieba)
	tiebaList.DELETE("/:pid/:fid/ignore", IgnoreTieba)
	tiebaList.POST("/:pid/:fid/reset", ResetTieba)
	tiebaList.POST("/:pid/sync", RefreshTiebaList, RateLimit(1, time.Second*10))
	tiebaList.GET("/status", GetForumStatus)
	tiebaList.GET("/status/:pid", GetForumStatus)

	// manage
	admin := api.Group("/admin", AdminCheck)
	admin.GET("/settings", GetAdminSettings)
	admin.POST("/settings", UpdateAdminSettings)
	admin.POST("/settings/:option/reset", ResetAdminSettings)
	admin.GET("/account", GetAccountsList)
	admin.DELETE("/account/:uid", AdminDeleteAccount)
	admin.PATCH("/account/:uid/modify", AdminModifyAccountInfo)
	admin.DELETE("/account/:uid/token", AdminDeleteAccountToken)
	admin.DELETE("/account/:uid/list", AdminDeleteTiebaAccountList)
	admin.POST("/account/:uid/list/reset", AdminResetTiebaList)
	admin.POST("/account/:uid/password/reset", AdminResetPassword)
	admin.POST("/account/:uid/plugin/:plugin_name/reset", AdminResetAccountPlugin)
	admin.POST("/service/push/mail/test", SendTestMessage)

	/// plugin
	admin.POST("/plugin/:plugin_name/switch", PluginSwitch)
	admin.DELETE("/plugin/:plugin_name", PluginUninstall)
	admin.GET("/plugin/:plugin_name/settings", GetPluginSettings)
	admin.POST("/plugin/:plugin_name/settings", UpdatePluginSettings)

	/// server
	admin.GET("/server/status", GetServerStatus)

	admin.POST("/server/encrypt", EncryptDB)
	// admin.POST("/server/decrypt", DecryptDB)

	if share.BuildPublishType == "binary" {
		admin.POST("/server/upgrade", UpgradeSystem, RateLimit(1, time.Second*10))
		admin.GET("/server/upgrade/releases", GetReleases, RateLimit(1, time.Second*10))
		if _function.VerifyPublicKey != nil {
			admin.POST("/server/upgrade/upload", UpgradeSystem2, RateLimit(1, time.Second*10), middleware.BodyLimit(bodyLimit50M))
		}
		admin.POST("/server/shutdown", ShutdownSystem)
	}

	// cron
	admin.GET("/server/cron", GetCronJobs)
	admin.POST("/server/cron/:id/run", RunCronJob)

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
		for _, r := range v.GetEndpoints() {
			plugin.Match([]string{r.Method}, "/"+v.GetInfo().Name+"/"+r.Path, r.Function)
		}
	}

	// frontend
	e.Any("/robots.txt", echoRobots)
	if share.EnableFrontend {
		fe, _ := fs.Sub(assets.EmbeddedFrontend, "dist")
		e.Any("/favicon.ico", echoFavicon)
		e.GET("/icp.jsonp", func(c *echo.Context) error {
			return c.JSONP(200, "__GetICP", ICPStruct{
				ICP: _function.GetOption("icp"),
			})
		})
		e.GET("/site.jsonp", func(c *echo.Context) error {
			feSettings := FESettings{
				SystemName:        _function.GetOption("system_name"),
				SystemKeywords:    _function.GetOption("system_keywords"),
				SystemDescription: _function.GetOption("system_description"),
				ICP:               _function.GetOption("icp"),
			}

			if share.DangerFrontend {
				feSettings.Footer = _function.GetOption("footer")
			}

			return c.JSONP(200, "__GetConfig", feSettings)
		})

		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				if !IsAPIPath(c) {
					// etag := _function.Sha256(c.Request().RequestURI)
					//
					// c.Response().Header().Set("ETag", `"`+etag+`"`)
					// if c.Request().Header.Get("If-None-Match") == `"`+etag+`"` {
					// 	return c.NoContent(http.StatusNotModified)
					// }

					c.Response().Header().Set("Last-Modified", share.BuildAtTime.UTC().Format(http.TimeFormat))
				}

				return next(c)
			}
		}, middleware.StaticWithConfig(middleware.StaticConfig{
			Filesystem: fe,
			HTML5:      true,
			Skipper:    IsAPIPath,
		}))
	}

	sc := echo.StartConfig{HideBanner: true}

	listener, _ := net.Listen(network, address)
	sc.Listener = listener

	e.Logger = slog.Default()

	if err := sc.Start(ctx, e); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}

func IsAPIPath(c *echo.Context) bool {
	path := c.Path()

	return slices.Contains(IndependentFEPath, path) || path == "/api" || strings.HasPrefix(path, "/api/")
}

type ICPStruct struct {
	ICP string `json:"icp"`
}

type FESettings struct {
	SystemName        string `json:"system_name"`
	SystemKeywords    string `json:"system_keywords"`
	SystemDescription string `json:"system_description"`
	Footer            string `json:"footer"`
	ICP               string `json:"icp"`
}
