package _api

import (
	"log"
	"net/http"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func ParsePath(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Path()
		method := c.Request().Method
		log.Println(method, path, c.Request().URL.Path, c.QueryString())

		if share.EnableFrontend && strings.HasPrefix(path, "/api/") {
			path = strings.TrimPrefix(path, "/api")
		}

		c.Set("path", path)

		return next(c)
	}
}

func SetHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !share.EnableFrontend {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Authorization")
		}

		if c.Request().Method == http.MethodOptions {
			c.Response().Header().Set("Access-Control-Max-Age", "86400")
			return c.NoContent(http.StatusNoContent)
		}

		// c.Response().Header().Add("X-Powered-By", "TbSignGo->")
		// c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
		return next(c)
	}
}

func AuthCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authorization := ""
		authSource := "header"
		if share.EnableFrontend {
			_authorization, err := c.Cookie("tc_auth")
			if err == nil {
				authorization = _authorization.Value
				authSource = "cookie"
			}
		}

		if authorization == "" {
			authorization = c.Request().Header.Get("Authorization")
		}

		uid, role := verifyAuthorization(authorization)

		// login
		if uid == _function.GuestUID || role == _function.RoleGuest {
			if authSource == "cookie" {
				c.SetCookie(&http.Cookie{
					Name:     "tc_auth",
					Value:    "",
					MaxAge:   -1,
					Expires:  time.Unix(0, 0),
					Path:     "/api",
					HttpOnly: true,
				})
			}
			return c.JSON(http.StatusOK, _function.ApiTemplate(401, "无效 session", _function.EchoEmptyObject, "tbsign"))
		}

		// deleted
		/// why this check here?
		if role == _function.RoleDeleted {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "账号已删除", _function.EchoEmptyObject, "tbsign"))
		}

		// banned
		if role == _function.RoleBanned {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "受限账号", _function.EchoEmptyObject, "tbsign"))
		}

		c.Set("uid", uid)
		c.Set("role", role)

		return next(c)
	}
}

func AdminCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// admin
		if role, ok := c.Get("role").(string); !ok || role != _function.RoleAdmin {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效用户组", _function.EchoEmptyObject, "tbsign"))
		}
		return next(c)
	}
}

func PluginPathPrecheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := strings.TrimPrefix(c.Get("path").(string), "/plugins/")

		pluginName := strings.SplitN(path, "/", 2)[0]

		_pluginInfo, ok := _plugin.PluginList[pluginName]

		if !ok || !_pluginInfo.(_plugin.PluginHooks).GetSwitch() {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "插件不可用", _function.EchoEmptyObject, "tbsign"))
		}

		return next(c)
	}
}

func RateLimit(_rate int, expiresIn time.Duration) echo.MiddlewareFunc {
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(_rate), Burst: 0, ExpiresIn: expiresIn},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusServiceUnavailable, _function.ApiTemplate(503, "服务不可用", _function.EchoEmptyObject, "tbsign"))
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, _function.ApiTemplate(429, "请求过多", _function.EchoEmptyObject, "tbsign"))
		},
	}

	return middleware.RateLimiterWithConfig(config)
}
