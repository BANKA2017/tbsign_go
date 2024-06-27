package _api

import (
	"net/http"
	"os"
	"runtime"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
)

func GetServerStatus(c echo.Context) error {
	hostname, _ := os.Hostname()

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"hostname":        hostname,
		"goroutine":       runtime.NumGoroutine(),
		"goversion":       runtime.Version(),
		"variables":       c.Get("variables"),
		"cron_sign_again": _function.GetOption("cron_sign_again"),
	}, "tbsign"))
}

func GetPluginsList(c echo.Context) error {
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", _function.PluginList, "tbsign"))
}

func GetLoginPageConfig(c echo.Context) error {
	// email
	enabledEmail := true

	mail := _function.GetOption("mail_name")
	mail_name := _function.GetOption("mail_yourname")
	smtp_host := _function.GetOption("mail_host")
	smtp_port := _function.GetOption("mail_port")
	smtp_secure := _function.GetOption("mail_secure")
	smtp_auth := _function.GetOption("mail_auth") != "0"
	smtp_username := _function.GetOption("mail_smtpname")
	smtp_password := _function.GetOption("mail_smtppw")

	if mail == "" || mail_name == "" || smtp_host == "" || smtp_port == "" || smtp_secure == "" {
		enabledEmail = false
	}

	if smtp_auth && (smtp_username == "" || smtp_password == "") {
		enabledEmail = false
	}

	// signup
	enabledSignup := _function.GetOption("enable_reg") != "0"
	closedCRegistrationMessage := ""
	if !enabledSignup {
		closedCRegistrationMessage = _function.GetOption("stop_reg")
	}
	// invite code
	enabledInviteCode := len(_function.GetOption("yr_reg")) > 0

	var resp = struct {
		EnabedResetPassword       bool   `json:"enabled_reset_password"`
		EnabledInviteCode         bool   `json:"enabled_invite_code"`
		EnabledSignup             bool   `json:"enabled_signup"`
		ClosedRegistrationMessage string `json:"closed_registration_message"`
	}{
		EnabedResetPassword:       enabledEmail,
		EnabledInviteCode:         enabledInviteCode,
		EnabledSignup:             enabledSignup,
		ClosedRegistrationMessage: closedCRegistrationMessage,
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}
