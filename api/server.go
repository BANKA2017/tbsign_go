package _api

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v4"
)

type PluginListContent struct {
	Name   string `json:"name"`
	Ver    string `json:"ver"`
	Status bool   `json:"status"`
}

func GetServerStatus(c echo.Context) error {
	//system
	//var memstats runtime.MemStats
	//runtime.ReadMemStats(&memstats)

	// count
	/// accounts
	var UIDCount int64
	_function.GormDB.R.Model(&model.TcUser{}).Count(&UIDCount)

	/// pid
	var PIDCount int64
	_function.GormDB.R.Model(&model.TcBaiduid{}).Count(&PIDCount)

	/// pid
	var ForumCount int64
	_function.GormDB.R.Model(&model.TcTieba{}).Count(&ForumCount)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"goroutine": runtime.NumGoroutine(),
		"goversion": fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		//"system":          fmt.Sprintf("cpu:%d, mem: [Alloc %d / Sys %d] MiB", runtime.NumCPU(), memstats.Alloc/1024/1024, memstats.Sys/1024/1024),
		"variables": c.Get("variables"),
		"build": map[string]string{
			"date":                          share.BuiltAt,
			"runtime":                       share.BuildRuntime,
			"commit_hash":                   share.BuildGitCommitHash,
			"embedded_frontend_commit_hash": share.BuildEmbeddedFrontendGitCommitHash,
		},
		"cron_sign_again": _function.GetOption("cron_sign_again"),
		"compat":          _function.GetOption("core_version"),
		"pure_go":         _function.GetOption("go_ver") == "1",
		"uid_count":       fmt.Sprintf("%d [online:%d]", UIDCount, len(keyBucket)),
		"pid_count":       PIDCount,
		"forum_count":     ForumCount,
	}, "tbsign"))
}

func GetPluginsList(c echo.Context) error {
	var resPluginList = make(map[string]PluginListContent)

	for k, v := range _function.PluginList {
		resPluginList[k] = PluginListContent{
			Name:   v.Name,
			Ver:    v.Ver,
			Status: v.Status,
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resPluginList, "tbsign"))
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
		SystemURL                 string `json:"system_url"`
	}{
		EnabedResetPassword:       enabledEmail,
		EnabledInviteCode:         enabledInviteCode,
		EnabledSignup:             enabledSignup,
		ClosedRegistrationMessage: closedCRegistrationMessage,
		SystemURL:                 _function.GetOption("system_url"),
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}
