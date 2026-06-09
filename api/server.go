package _api

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/kdnetwork/code-snippet/go/db"
	"github.com/kdnetwork/code-snippet/go/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/singleflight"
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
	checkinStatus := new(_type.StatusStruct)

	today := strconv.Itoa(time.Now().Day())
	_function.GormDB.R.Model(&model.TcTieba{}).Select("SUM(CASE WHEN (no = 0) AND status = 0 AND latest = ? THEN 1 ELSE 0 END) AS success, SUM(CASE WHEN (no = 0) AND status <> 0 AND latest = ? THEN 1 ELSE 0 END) AS failed, SUM(CASE WHEN (no = 0) AND latest <> ? THEN 1 ELSE 0 END) AS waiting, SUM(CASE WHEN no <> 0 THEN 1 ELSE 0 END) AS is_ignore", today, today, today).Scan(checkinStatus)

	ForumCount := checkinStatus.Success + checkinStatus.Failed + checkinStatus.Waiting + checkinStatus.IsIgnore

	// vcs, ok := debug.ReadBuildInfo()
	// if !ok {
	// 	slog.Error("failed to read build info")
	// }

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"goroutine":  runtime.NumGoroutine(),
		"goversion":  fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		"start_time": share.StartTime.UnixMilli(),
		//"system":          fmt.Sprintf("cpu:%d, mem: [Alloc %d / Sys %d] MiB", runtime.NumCPU(), memstats.Alloc/1024/1024, memstats.Sys/1024/1024),
		"variables": map[string]any{
			"dbmode":    share.DBMode,
			"dbversion": share.DBVersion,
			"tlsdb":     share.DBTLSOption != "false" && share.DBTLSOption != "",
			"testmode":  share.TestMode,
		},
		"build": map[string]string{
			"date":                          share.BuiltAt,
			"runtime":                       share.BuildRuntime,
			"commit_hash":                   share.BuildGitCommitHash,
			"embedded_frontend_commit_hash": share.BuildEmbeddedFrontendGitCommitHash,
			"publish_type":                  share.BuildPublishType,
			"cgo":                           _function.When(db.CgoEnabled, "1", "0"),
			"vcs.modified":                  _function.When(share.BuildDirty, "1", "0"),
			// "vcs":                           vcs,
		},
		"upgrade": map[string]any{
			"api_base":     share.ReleaseApiBase,
			"asset_base":   _function.ReleaseFilesPath,
			"allow_upload": _function.VerifyPublicKey != nil,
		},
		"cron_sign_again": _function.GetOption("cron_sign_again"),
		"compat":          _function.GetOption("core_version"),
		"pure_go":         share.IsPureGO,
		"encrypt":         share.IsEncrypt,
		"uid_count":       strconv.Itoa(int(UIDCount)),
		"pid_count":       PIDCount,
		"forum_count":     ForumCount,
		"checkin_status":  checkinStatus,
	}, "tbsign"))
}

var upgradeSF singleflight.Group

func UpgradeSystem(c echo.Context) error {
	_, err, _ := upgradeSF.Do("upgrade", func() (any, error) {
		version := c.FormValue("version")
		var err error

		if _function.GetOption("go_next_upgrade_func") == "1" {
			err = _function.Upgrade2("tbsign_go." + strings.TrimSpace(version))
		} else {
			err = _function.Upgrade(strings.TrimSpace(version))
		}

		if err != nil {
			return nil, c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, err.Error(), map[string]any{}, "tbsign"))
		}

		return nil, ShutdownSystem(c)
	})
	return err
}

func readFormFile(c echo.Context, name string) ([]byte, error) {
	file, err := c.FormFile(name)
	if err != nil {
		return nil, err
	}
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	return io.ReadAll(src)
}

func UpgradeSystem2(c echo.Context) error {
	_, err, _ := upgradeSF.Do("upgrade", func() (any, error) {
		metadata, err := readFormFile(c, "metadata")
		if err != nil {
			return nil, c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, err.Error(), map[string]any{}, "tbsign"))
		}

		bin, err := readFormFile(c, "binary")
		if err != nil {
			return nil, c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, err.Error(), map[string]any{}, "tbsign"))
		}

		if err = _function.Upgrade3(bin, string(metadata)); err != nil {
			return nil, c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, err.Error(), map[string]any{}, "tbsign"))
		}

		return nil, ShutdownSystem(c)
	})
	return err
}

func ShutdownSystem(c echo.Context) error {
	go func() {
		slog.Info("system.shutdown")
		defer os.Exit(1)
		<-time.After(time.Second)
	}()
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{}, "tbsign"))
}

func GetPluginsList(c echo.Context) error {
	isAdmin := c.Get("role").(string) == _function.RoleAdmin
	var resPluginList = make(map[string]*PluginListContent, len(_plugin.PluginList))

	for name, info := range _plugin.PluginList {
		value := info.GetInfo()
		status := _function.TinyIntToBool(value.Info.Status)

		if !isAdmin && !status {
			continue
		}

		resPluginList[name] = _function.VPtr(PluginListContent{
			Name:   value.Name,
			Ver:    value.Info.Ver,
			Status: status,

			PluginNameCN:      value.PluginNameCN,
			PluginNameCNShort: value.PluginNameCNShort,
			PluginNameFE:      value.PluginNameFE,
		})
		if isAdmin {
			settingOptions := make([]PluginListSettingOption, 0, len(value.SettingOptions))

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

type CronJob struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Tags            []string `json:"tags"`
	LastStartAt     int64    `json:"last_start_at"`
	LastCompletedAt int64    `json:"last_completed_at"`
	NextTime        int64    `json:"next_time"`
	Running         bool     `json:"running"`
}

var cronJobOrder = map[string]int{
	"tc_service": 1,
	"checkin":    2,
	"report":     3,
	"plugin":     4,
}

func GetCronJobs(c echo.Context) error {
	var cronJobs = _function.Crontab.Jobs()

	var CronJobList = make([]CronJob, 0, len(cronJobs))

	for _, job := range cronJobs {
		lastStart, _ := job.LastRunStartedAt()
		lastCompleted, _ := job.LastRunCompletedAt()
		nextTime, _ := job.NextRun()
		running, _ := job.IsRunning()

		CronJobList = append(CronJobList, CronJob{
			ID:              job.ID().String(),
			Name:            job.Name(),
			Tags:            job.Tags(),
			LastStartAt:     utils.Clamp(lastStart.Unix(), -1, math.MaxInt64),
			LastCompletedAt: utils.Clamp(lastCompleted.Unix(), -1, math.MaxInt64),
			NextTime:        utils.Clamp(nextTime.Unix(), -1, math.MaxInt64),
			Running:         running,
		})
	}

	sort.Slice(CronJobList, func(i, j int) bool {
		a := CronJobList[i]
		b := CronJobList[j]

		aMulti := len(CronJobList[i].Tags) >= 2
		bMulti := len(CronJobList[j].Tags) >= 2

		if aMulti != bMulti {
			return bMulti
		}

		if aMulti {
			return a.Tags[1] < b.Tags[1]
		}

		return cronJobOrder[a.Tags[0]] < cronJobOrder[b.Tags[0]]
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", CronJobList, "tbsign"))
}

func RunCronJob(c echo.Context) error {
	jobID := c.Param("id")

	jobs := _function.Crontab.Jobs()
	for _, job := range jobs {
		if job.ID().String() == jobID {
			if err := job.RunNow(); err != nil {
				slog.Error("failed to run cron job", "id", jobID, "error", err)
				return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, err.Error(), false, "tbsign"))
			}
			return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
		}
	}

	return c.JSON(http.StatusNotFound, _function.ApiTemplate(404, "Cron job not found", false, "tbsign"))
}
