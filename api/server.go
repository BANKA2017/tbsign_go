package _api

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v4"
)

type PluginListSettingOption struct {
	OptionName   string `json:"option_name"`
	OptionNameCN string `json:"option_name_cn"`
}

type PluginListContent struct {
	Name   string `json:"name"`
	Ver    string `json:"ver"`
	Status bool   `json:"status"`

	PluginNameCN      string `json:"plugin_name_cn"`
	PluginNameCNShort string `json:"plugin_name_cn_short"`
	PluginNameFE      string `json:"plugin_name_fe"`

	SettingOptions []PluginListSettingOption `json:"setting_options,omitempty"`
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

	/// forums
	checkinStatus := new(struct {
		Success  int64 `json:"success"`
		Failed   int64 `json:"failed"`
		Waiting  int64 `json:"waiting"`
		IsIgnore int64 `json:"ignore"`
	})

	today := strconv.Itoa(_function.Now.Local().Day())
	_function.GormDB.R.Model(&model.TcTieba{}).Select("SUM(CASE WHEN NOT (no = 0) AND status = 0 AND latest = "+today+" THEN 1 ELSE 0 END) AS success", "SUM(CASE WHEN NOT (no = 0) AND status <> 0 AND latest = "+today+" THEN 1 ELSE 0 END) AS failed", "SUM(CASE WHEN NOT (no = 0) AND latest <> "+today+" THEN 1 ELSE 0 END) AS waiting", "SUM(CASE WHEN no <> 0 THEN 1 ELSE 0 END) AS is_ignore").Scan(checkinStatus)

	ForumCount := checkinStatus.Success + checkinStatus.Failed + checkinStatus.Waiting + checkinStatus.IsIgnore

	onlineCount := 0
	HttpAuthRefreshTokenMap.Range(func(key, value any) bool {
		onlineCount++
		return true
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"goroutine":  runtime.NumGoroutine(),
		"goversion":  fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		"start_time": share.StartTime.UnixMilli(),
		//"system":          fmt.Sprintf("cpu:%d, mem: [Alloc %d / Sys %d] MiB", runtime.NumCPU(), memstats.Alloc/1024/1024, memstats.Sys/1024/1024),
		"variables": c.Get("variables"),
		"build": map[string]string{
			"date":                          share.BuiltAt,
			"runtime":                       share.BuildRuntime,
			"commit_hash":                   share.BuildGitCommitHash,
			"embedded_frontend_commit_hash": share.BuildEmbeddedFrontendGitCommitHash,
			"publish_type":                  share.BuildPublishType,
		},
		"cron_sign_again": _function.GetOption("cron_sign_again"),
		"compat":          _function.GetOption("core_version"),
		"pure_go":         _function.GetOption("go_ver") == "1",
		"uid_count":       fmt.Sprintf("%d [online:%d]", UIDCount, onlineCount),
		"pid_count":       PIDCount,
		"forum_count":     ForumCount,
		"checkin_status":  checkinStatus,
	}, "tbsign"))
}

func UpgradeSystem(c echo.Context) error {
	version := c.FormValue("version")
	err := _function.Upgrade(strings.TrimSpace(version))

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, err.Error(), map[string]any{}, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{}, "tbsign"))
}

func ShutdownSystem(c echo.Context) error {
	os.Exit(1)
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{}, "tbsign"))
}

func GetPluginsList(c echo.Context) error {
	isAdmin := c.Get("role").(string) == "admin"
	var resPluginList = make(map[string]*PluginListContent)

	for name, info := range _plugin.PluginList {
		value := info.(_plugin.PluginHooks).GetInfo()

		resPluginList[name] = _function.VariablePtrWrapper(PluginListContent{
			Name:   value.Name,
			Ver:    value.Info.Ver,
			Status: _function.TinyIntToBool(value.Info.Status),

			PluginNameCN:      value.PluginNameCN,
			PluginNameCNShort: value.PluginNameCNShort,
			PluginNameFE:      value.PluginNameFE,
		})
		if isAdmin {
			settingOptions := []PluginListSettingOption{}

			for _, settingOptionsItem := range value.SettingOptions {
				settingOptions = append(settingOptions, PluginListSettingOption{
					OptionName:   settingOptionsItem.OptionName,
					OptionNameCN: settingOptionsItem.OptionNameCN,
				})
			}
			resPluginList[name].SettingOptions = settingOptions
		}
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resPluginList, "tbsign"))
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
		EnabedEmail               bool   `json:"enabled_email"`
		EnabledInviteCode         bool   `json:"enabled_invite_code"`
		EnabledSignup             bool   `json:"enabled_signup"`
		ClosedRegistrationMessage string `json:"closed_registration_message"`
		SystemURL                 string `json:"system_url"`
	}{
		EnabedEmail:               enabledEmail,
		EnabledInviteCode:         enabledInviteCode,
		EnabledSignup:             enabledSignup,
		ClosedRegistrationMessage: closedCRegistrationMessage,
		SystemURL:                 _function.GetOption("system_url"),
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp, "tbsign"))
}
