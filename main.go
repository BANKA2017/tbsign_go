package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	_api "github.com/BANKA2017/tbsign_go/api"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/kdnetwork/code-snippet/go/db"
	"github.com/kdnetwork/code-snippet/go/utils"
	"gorm.io/gorm/logger"
)

// install
var setup bool
var autoInstall bool

var adminName string
var adminEmail string
var adminPassword string

var EncryptDataAction string

var err error

func init() {
	_function.InitLogger()
	fmt.Println("TbSign➡️\n--- info ---")
	fmt.Println("build_at:", share.BuiltAt)
	fmt.Println("build_runtime:", runtime.Version()+" "+share.BuildRuntime)
	fmt.Println("commit_hash:", share.BuildGitCommitHash)
	fmt.Println("frontend_hash:", share.BuildEmbeddedFrontendGitCommitHash)
	fmt.Println("release_type:", share.BuildPublishType)
	if share.BuiltAt != "Now" && share.BuildGitCommitHash != "N/A" && share.BuildEmbeddedFrontendGitCommitHash != "N/A" {
		fmt.Println("version:", fmt.Sprintf("%s.%s.%s\n", share.BuildAtTime.Format("20060102"), share.BuildGitCommitHash[0:7], share.BuildEmbeddedFrontendGitCommitHash[0:7]))
	} else {
		fmt.Println("version: dev\n------------")
	}
}

var (
	buildInReleaseFileBasePath = share.ReleaseFilesPath
)

func main() {
	// sqlite
	flag.StringVar(&share.DBPath, "db_path", utils.GetEnv("tc_db_path", ""), "Database path")

	// mysql
	var tmpHost string
	flag.StringVar(&share.DBUsername, "username", utils.GetEnv("tc_username", ""), "Username")
	flag.StringVar(&share.DBPassword, "pwd", utils.GetEnv("tc_pwd", ""), "Password")
	flag.StringVar(&tmpHost, "endpoint", utils.GetEnv("tc_endpoint", "127.0.0.1:3306"), "MySQL host:port (deprecated)")
	flag.StringVar(&share.DBEndpoint, "host", utils.GetEnv("tc_host", "127.0.0.1:3306"), "MySQL host:port")
	flag.StringVar(&share.DBName, "db", utils.GetEnv("tc_db", "tbsign"), "Database name")
	flag.StringVar(&share.DBMode, "db_mode", utils.GetEnv("tc_db_mode", "mysql"), "sqlite/mysql/(pgsql|postgresql)")
	flag.StringVar(&share.DBTLSOption, "db_tls", utils.GetEnv("tc_db_tls", "false"), "Option for CA cert (MySQL/PostgreSQL)")

	//proxy
	flag.BoolVar(&_function.IgnoreProxy, "no_proxy", false, "Ignore the http proxy config from environment vars")

	// api
	flag.BoolVar(&share.EnableApi, "api", utils.GetEnv("tc_api", "") != "", "active backend endpoints")
	flag.BoolVar(&share.EnableFrontend, "fe", utils.GetEnv("tc_fe", "") != "", "active frontend endpoints")
	flag.BoolVar(&share.EnableBackup, "allow_backup", utils.GetEnv("tc_allow_backup", "") != "", "allow backup (export/import)")
	flag.StringVar(&share.Address, "address", utils.GetEnv("tc_address", ":1323"), "address :1323")

	// setup
	flag.BoolVar(&setup, "setup", false, "Init the system [force]")
	flag.BoolVar(&autoInstall, "auto_install", utils.GetEnv("tc_auto_install", "") != "", "Auto install the system when tables are not exist")
	flag.StringVar(&adminName, "admin_name", utils.GetEnv("tc_admin_name", ""), "Name of admin")
	flag.StringVar(&adminEmail, "admin_email", utils.GetEnv("tc_admin_email", ""), "Email of admin")
	flag.StringVar(&adminPassword, "admin_password", utils.GetEnv("tc_admin_password", ""), "Password of admin")

	// --experimental-*
	// encrypt
	flag.StringVar(&EncryptDataAction, "data_encrypt_action", "", "Encrypt/Decrypt data in database")
	flag.StringVar(&share.DataEncryptKeyStr, "data_encrypt_key", utils.GetEnv("tc_data_encrypt_key", ""), "The key to encrypt some user data (base64url)")
	// flag.BoolVar(&share.DisableEmail, "disable-email", false, "disable email")
	flag.StringVar(&share.DNSAddress, "dns_addr", utils.GetEnv("tc_dns_addr", ""), "DNS Address")
	// releases
	flag.StringVar(&share.ReleaseFilesPath, "release_file_base", utils.GetEnv("tc_release_file_base", buildInReleaseFileBasePath), "Base path for release files")
	flag.StringVar(&share.ReleaseApiBase, "release_api_base", utils.GetEnv("tc_release_api_base", share.ReleaseApiBase), "Base path for release API")

	// others
	flag.BoolVar(&share.TestMode, "test", utils.GetEnv("tc_test", "") != "", "Not send any requests to tieba servers")

	flag.Parse()

	if share.ReleaseFilesPath != buildInReleaseFileBasePath {
		_function.ReleaseFilesPath = share.ReleaseFilesPath
	}

	if share.DBEndpoint == "127.0.0.1:3306" && tmpHost != "127.0.0.1:3306" {
		share.DBEndpoint = tmpHost
	}

	if share.DBTLSOption == "false" && (share.DBMode == "pgsql" || share.DBMode == db.DBModePostgreSQL) {
		share.DBTLSOption = "prefer" // default value
	}

	if share.DBPath == "" && share.DBEndpoint == "" {
		_function.Fatal("无效数据库")
	}

	if !share.EnableApi && share.EnableFrontend {
		_function.Fatal("不允许关闭 api 的同时又启用前端!!!")
	}

	if share.DataEncryptKeyStr != "" {
		share.DataEncryptKeyByte, err = _function.Base64URLDecode(share.DataEncryptKeyStr)
		if err != nil {
			_function.Fatal("密钥不可用", "error", err)
		}
		if len(share.DataEncryptKeyByte) != 32 {
			_function.Fatal("密钥长度无效")
		}
	}

	if setup {
		slog.Warn("覆盖安装已启用，会覆盖现有数据，请做好备份")
	}

	if setup && autoInstall {
		_function.Fatal("不允许自动化覆盖安装!!!")
	} else if autoInstall && adminName != "" && adminEmail != "" && adminPassword != "" {
		slog.Warn("已启用自动安装")
	} else if autoInstall {
		_function.Fatal("管理员信息不完整，无法安装")
	}

	// TODO setup slog

	logLevel := logger.Error
	// slogLevel := slog.LevelError
	if share.TestMode {
		logLevel = logger.Info
		_function.SlogLevel = slog.LevelDebug
	}

	_function.InitLogger()

	// connect to db
	dbExists := true

	_function.GormDB.LogLevel = logLevel
	_function.GormDB.ServicePrefix = "tbsign-db"
	_function.GormDB.WALMode = true
	_function.GormDB.SetDialTimeout(_function.VPtr(time.Second * 30))

	if share.DBMode == "pgsql" || share.DBMode == db.DBModePostgreSQL {
		// precheck table
		if err = _function.GormDB.SetDBMode(db.DBModePostgreSQL).SetDBAuth(share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, share.DBTLSOption).ConnectToDefault(); err != nil {
			_function.Fatal("db", "error", err)
		}

		dbExists, err = _function.GormDB.FastDBCheck(share.DBName)
		if err != nil {
			_function.Fatal("db", "error", err)
		} else if !dbExists {
			slog.Warn("db", "dbname", share.DBName, "status", "not exists")
			setup = true
		}

		// setup
		if setup {
			if !_plugin.SetupSystem(dbExists, autoInstall, adminName, adminEmail, adminPassword) {
				os.Exit(1)
			}
		} else if share.DBName != "postgres" {
			if err = _function.GormDB.Close(); err != nil {
				_function.Fatal("db.close", "error", err)
			}

			if err = _function.GormDB.Connect(); err != nil {
				_function.Fatal("db", "error", err)
			}
		}
	} else if share.DBPath != "" {
		// sqlite
		dbStat, err := os.Stat(share.DBPath)
		if err == nil {
			if dbStat.IsDir() {
				_function.Fatal("db.sqlite", "error", share.DBPath+" is a directory")
			}
		} else if errors.Is(err, fs.ErrNotExist) {
			dbExists = false
		} else {
			_function.Fatal("db", "error", err)
		}

		setup = setup || !dbExists

		if err = _function.GormDB.SetDBMode(db.DBModeSQLite).SetDBPath(share.DBPath).Connect(); err != nil {
			_function.Fatal("db", "error", err)
		}

		share.DBMode = _function.GormDB.DBMode

		// setup
		if setup {
			if !_plugin.SetupSystem(dbExists, autoInstall, adminName, adminEmail, adminPassword) {
				os.Exit(1)
			}
		}
	} else {
		// mysql
		if share.DBUsername == "" {
			_function.Fatal("Empty MySQL username")
		}
		// precheck table

		// hook for tls
		tlsOptionIsTrue := strings.EqualFold(share.DBTLSOption, "true")

		_function.GormDB.SetDBMode(db.DBModeMySQL).SetDBAuth(share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, _function.When(tlsOptionIsTrue, "", share.DBTLSOption))

		if tlsOptionIsTrue {
			_function.GormDB.SetCertPool(_function.CACertPool)
		}

		if err = _function.GormDB.ConnectToDefault(); err != nil {
			_function.Fatal("db", "error", err)
		}
		share.DBMode = _function.GormDB.DBMode

		dbExists, err = _function.GormDB.FastDBCheck(share.DBName)
		if err != nil {
			_function.Fatal("db", "error", err)
		} else if !dbExists {
			slog.Warn("db", "dbname", share.DBName, "status", "not exists")
			setup = true
		}

		// setup
		if setup {
			if !_plugin.SetupSystem(dbExists, autoInstall, adminName, adminEmail, adminPassword) {
				os.Exit(1)
			}
		} else {
			if err = _function.GormDB.Close(); err != nil {
				_function.Fatal("db.close", "error", err)
			}

			if err = _function.GormDB.Connect(); err != nil {
				_function.Fatal("db", "error", err)
			}
		}
	}

	adminEmail = ""
	adminName = ""
	adminPassword = ""

	// db version
	share.DBVersion = _function.GormDB.Version()

	// init
	_function.InitOptions()
	share.IsPureGO = _function.GetOption("go_ver") == "1"
	share.IsEncrypt = _function.GetOption("go_encrypt") != "0"
	_plugin.InitPluginList()

	// encrypt/decrypt init
	if share.IsPureGO && EncryptDataAction != "" {
		if len(share.DataEncryptKeyByte) != 32 {
			_function.FmtFatal("❌无效密钥，无法处理加密内容")
		} else if strings.EqualFold(EncryptDataAction, "encrypt") && len(share.DataEncryptKeyByte) > 0 {
			err := _plugin.EncryptBaiduIDData()
			if err != nil {
				_function.FmtFatal("❌crypto.encrypt", err)
			}
			fmt.Println("✅crypto.encrypt: 加密完成")
			os.Exit(0)
		} else if strings.EqualFold(EncryptDataAction, "decrypt") && len(share.DataEncryptKeyByte) > 0 {
			err := _plugin.DecryptBaiduIDData()
			if err != nil {
				_function.FmtFatal("❌crypto.decrypt", err)
			}
			fmt.Println("✅crypto.decrypt: 解密完成")
			os.Exit(0)
		}
	} else if len(share.DataEncryptKeyByte) == 32 && !share.IsPureGO {
		// DO NOT USE ENCRYPT IN COMPAT MODE!!!
		share.DataEncryptKeyByte = []byte{}
		share.DataEncryptKeyStr = ""

		slog.Warn("兼容模式下不支持加密，已恢复使用明文")
	}

	if share.IsEncrypt && len(share.DataEncryptKeyByte) != 32 {
		_function.Fatal("无效密钥，无法处理加密内容")
	}
	if !share.IsEncrypt && len(share.DataEncryptKeyByte) > 0 {
		share.DataEncryptKeyByte = []byte{}
		share.DataEncryptKeyStr = ""

		slog.Warn("数据未加密，已恢复使用明文")
	}

	/// client
	/// DO NOT EXEC _function.InitClient BEFORE READING FLAGS AND ENV!!!!!
	_function.DefaultClient = _function.InitClient(30 * time.Minute)
	_function.TBClient = _function.InitClient(10 * time.Second)

	if share.EnableApi {
		go _api.Api(share.Address)
	}

	// wait for the next minute
	if _function.GetOption("go_wait_for_next_minute_on_startup") != "0" {
		now := time.Now()
		nextMinute := now.Truncate(time.Minute).Add(time.Minute)
		slog.Info("等待到下一个整分钟以启动", "next_minute", nextMinute)
		time.Sleep(nextMinute.Sub(now))
	}

	// Interval
	oneMinuteInterval := time.NewTicker(time.Minute)
	defer oneMinuteInterval.Stop()
	fourHoursInterval := time.NewTicker(time.Hour * 4)
	defer fourHoursInterval.Stop()

	// cron
	for {
		select {
		case <-oneMinuteInterval.C:
			if share.TestMode {
				if share.CrontabBypassTimes.Load() > 0 {
					share.CrontabBypassTimes.Add(-1)
				} else {
					continue
				}
			}
			_plugin.DoCheckinAction()
			_plugin.DoReCheckinAction()

			// plugins
			for _, info := range _plugin.PluginList {
				if _function.TinyIntToBool(info.GetDBInfo().Status) {
					go info.Action()
				}
			}

			// daily report
			_plugin.DailyReportAction()
		case <-fourHoursInterval.C:
			_function.InitOptions()
			_plugin.InitPluginList()
		}
	}
}
