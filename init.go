package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/go-co-op/gocron/v2"
	"github.com/kdnetwork/code-snippet/go/db"
	"github.com/kdnetwork/code-snippet/go/log"
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

func InitEnv() {
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
	flag.BoolVar(&share.EnableApi, "api", utils.GetBoolEnv("tc_api"), "active backend endpoints")
	flag.BoolVar(&share.EnableFrontend, "fe", utils.GetBoolEnv("tc_fe"), "active frontend endpoints")
	flag.BoolVar(&share.EnableBackup, "allow_backup", utils.GetBoolEnv("tc_allow_backup"), "allow backup (export/import)")
	flag.StringVar(&share.Address, "address", utils.GetEnv("tc_address", ":1323"), "address :1323")
	flag.StringVar(&share.Network, "network", utils.GetEnv("tc_network", "tcp"), "tcp/tcp4/tcp6/unix/unixpacket")

	// setup
	flag.BoolVar(&setup, "setup", false, "Init the system [force]")
	flag.BoolVar(&autoInstall, "auto_install", utils.GetBoolEnv("tc_auto_install"), "Auto install the system when tables are not exist")
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
	flag.StringVar(&share.ReleaseFilesPath, "release_file_base", utils.GetEnv("tc_release_file_base", share.ReleaseFilesPath), "Base path for release files")
	flag.StringVar(&share.ReleaseApiBase, "release_api_base", utils.GetEnv("tc_release_api_base", share.ReleaseApiBase), "To replace `https://api.github.com/repos/{owner}/{repo}`")

	// others
	flag.BoolVar(&share.TestMode, "test", utils.GetBoolEnv("tc_test"), "Not send any requests to tieba servers")
	flag.BoolVar(&share.DangerFrontend, "danger_fe", utils.GetBoolEnv("tc_danger_fe"), "Allow dangerous frontend actions")

	flag.Parse()

	if share.DBEndpoint == "127.0.0.1:3306" && tmpHost != "127.0.0.1:3306" {
		share.DBEndpoint = tmpHost
	}

	if share.DBTLSOption == "false" && (share.DBMode == "pgsql" || share.DBMode == db.DBModePostgreSQL) {
		share.DBTLSOption = "prefer" // default value
	}

	if share.DBPath == "" && share.DBEndpoint == "" {
		log.Fatal("无效数据库")
	}

	if !share.EnableApi && share.EnableFrontend {
		log.Fatal("不允许关闭 api 的同时又启用前端!!!")
	}

	if share.DataEncryptKeyStr != "" {
		share.DataEncryptKeyByte, err = _function.Base64URLDecode(share.DataEncryptKeyStr)
		if err != nil {
			log.Fatal("密钥不可用", "error", err)
		}
		if len(share.DataEncryptKeyByte) != 32 {
			log.Fatal("密钥长度无效")
		}
	}

	if setup {
		slog.Warn("覆盖安装已启用，会覆盖现有数据，请做好备份")
	}

	if setup && autoInstall {
		log.Fatal("不允许自动化覆盖安装!!!")
	} else if autoInstall && adminName != "" && adminEmail != "" && adminPassword != "" {
		slog.Warn("已启用自动安装")
	} else if autoInstall {
		log.Fatal("管理员信息不完整，无法安装")
	}

	logLevel := logger.Error
	// slogLevel := slog.LevelError
	if share.TestMode {
		logLevel = logger.Info
		_function.SlogLevel.Set(slog.LevelDebug)
	}

	_function.GormDB.LogLevel = logLevel
}

func InitDB() {
	dbExists := true

	_function.GormDB.ServicePrefix = "tbsign-db"
	_function.GormDB.WALMode = true
	_function.GormDB.SetDialTimeout(_function.VPtr(time.Second * 30)).SetLogger(logger.NewSlogLogger(
		slog.Default(), logger.Config{
			// copy from gorm default logger config
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  _function.GormDB.LogLevel,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		},
	))

	if share.DBMode == "pgsql" || share.DBMode == db.DBModePostgreSQL {
		// precheck table
		if err = _function.GormDB.SetDBMode(db.DBModePostgreSQL).SetDBAuth(share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, share.DBTLSOption).ConnectToDefault(); err != nil {
			log.Fatal("db", "error", err)
		}

		dbExists, err = _function.GormDB.FastDBCheck(share.DBName)
		if err != nil {
			log.Fatal("db", "error", err)
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
				log.Fatal("db.close", "error", err)
			}

			if err = _function.GormDB.Connect(); err != nil {
				log.Fatal("db", "error", err)
			}
		}
	} else if share.DBPath != "" {
		// sqlite
		dbStat, err := os.Stat(share.DBPath)
		if err == nil {
			if dbStat.IsDir() {
				log.Fatal("db.sqlite", "error", share.DBPath+" is a directory")
			}
		} else if errors.Is(err, fs.ErrNotExist) {
			dbExists = false
		} else {
			log.Fatal("db", "error", err)
		}

		setup = setup || !dbExists

		if err = _function.GormDB.SetDBMode(db.DBModeSQLite).SetDBPath(share.DBPath).Connect(); err != nil {
			log.Fatal("db", "error", err)
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
			log.Fatal("Empty MySQL username")
		}
		// precheck table

		// hook for tls
		tlsOptionIsTrue := strings.EqualFold(share.DBTLSOption, "true")

		_function.GormDB.SetDBMode(db.DBModeMySQL).SetDBAuth(share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, _function.When(tlsOptionIsTrue, "", share.DBTLSOption))

		if tlsOptionIsTrue {
			_function.GormDB.SetCertPool(_function.CACertPool)
		}

		if err = _function.GormDB.ConnectToDefault(); err != nil {
			log.Fatal("db", "error", err)
		}
		share.DBMode = _function.GormDB.DBMode

		dbExists, err = _function.GormDB.FastDBCheck(share.DBName)
		if err != nil {
			log.Fatal("db", "error", err)
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
				log.Fatal("db.close", "error", err)
			}

			if err = _function.GormDB.Connect(); err != nil {
				log.Fatal("db", "error", err)
			}
		}
	}

	adminEmail = ""
	adminName = ""
	adminPassword = ""

	// db version
	share.DBVersion = _function.GormDB.Version()
}

func InitCrontab() (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler(
	// gocron.WithLogger(
	// 	gocron.NewLogger(gocron.LogLevelDebug),
	// ),
	)
	if err != nil {
		return s, err
	}

	// self-services
	if _, err = s.NewJob(
		gocron.DurationJob(time.Hour*4),
		gocron.NewTask(func() {
			_function.InitOptions()
			_plugin.InitPluginList()
		}),
		gocron.WithTags("tc_service"),
		gocron.WithName("内部服务"),
	); err != nil {
		return s, err
	}

	// test mode
	var minuteDuration = gocron.DurationJob(time.Minute)
	if share.TestMode {
		minuteDuration = gocron.DurationJob(time.Hour * 24 * 99999)
	}

	// daily report
	if _, err = s.NewJob(
		minuteDuration,
		gocron.NewTask(_plugin.DailyReportAction),
		gocron.WithTags("report"),
		gocron.WithName("签到报告"),
	); err != nil {
		return s, err
	}

	// check-in actions
	if _, err = s.NewJob(
		minuteDuration,
		gocron.NewTask(func() {
			_plugin.DoCheckinAction()
			_plugin.DoReCheckinAction()
		}),
		gocron.WithTags("checkin"),
		gocron.WithName("签到任务"),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	); err != nil {
		return s, err
	}

	// plugins
	for _, plugin := range _plugin.PluginList {
		var d gocron.JobDefinition
		if plugin.GetInfo().RandomDuration && !share.TestMode {
			d = gocron.DurationRandomJob(time.Second*50, time.Second*70)
		} else {
			d = minuteDuration
		}

		if _, err = s.NewJob(
			d,
			gocron.NewTask(func() {
				if plugin.GetSwitch() {
					plugin.Action()
				}
			}),
			gocron.WithTags("plugin", "plugin:"+plugin.GetInfo().Name),
			gocron.WithName("插件:"+plugin.GetInfo().PluginNameCN),
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
			// gocron.WithIntervalFromCompletion(),
		); err != nil {
			return s, err
		}
	}

	return s, err
}

func InitEncrypt() {
	isValidEncryptKeyLength := len(share.DataEncryptKeyByte) == 32
	EncryptDataAction = strings.ToLower(EncryptDataAction)
	encryptStatus := _function.GetOption("go_encrypt")
	isDBEncrypted := encryptStatus != "0" && encryptStatus != ""

	// plaintext mode
	if !isDBEncrypted && len(share.DataEncryptKeyByte) == 0 {
		return
	}

	// key format
	if !isValidEncryptKeyLength && len(share.DataEncryptKeyByte) > 0 {
		share.DataEncryptKeyByte = []byte{}
		share.DataEncryptKeyStr = ""

		slog.Warn("无效密钥，无法处理加密内容，已恢复使用明文")
	} else if isValidEncryptKeyLength && !share.IsPureGO {
		// DO NOT USE ENCRYPT IN COMPAT MODE!!!
		share.DataEncryptKeyByte = []byte{}
		share.DataEncryptKeyStr = ""

		isValidEncryptKeyLength = false

		if EncryptDataAction != "" {
			log.FmtFatal("❌兼容模式下不支持加密操作")
		} else {
			slog.Warn("兼容模式下不支持加密，已恢复使用明文")
		}
	}

	// encrypt/decrypt action
	if EncryptDataAction != "" {
		switch EncryptDataAction {
		case "encrypt":
			if isDBEncrypted {
				log.FmtFatal("❌crypto.encrypt", errors.New("database already encrypted"))
			}

			err := _plugin.EncryptTCData()
			if err != nil {
				log.FmtFatal("❌crypto.encrypt", err)
			}

			fmt.Println("✅crypto.encrypt: 加密完成")
			os.Exit(0)

		case "decrypt":
			if err := _function.VerifyEncStatus(); err != nil {
				log.FmtFatal("❌crypto.decrypt", err)
			}

			err := _plugin.DecryptTCData()
			if err != nil {
				log.FmtFatal("❌crypto.decrypt", err)
			}

			fmt.Println("✅crypto.decrypt: 解密完成")
			os.Exit(0)
		default:
			log.FmtFatal("❌无效行为，无法处理加密内容")
		}
	}

	if isDBEncrypted {
		if !isValidEncryptKeyLength {
			log.Fatal("没有密钥或密钥无效", "error", errors.New("encrypt key required"))
		}

		if err := _function.VerifyEncStatus(); err != nil {
			log.Fatal("加密密钥错误", "error", err)
		}

		share.IsEncrypt = true

	} else if isValidEncryptKeyLength && setup {
		// first time
		if err := _function.CreateVerifyEncStatus(); err != nil {
			log.Fatal("初始化加密状态失败", "error", err)
		}

		share.IsEncrypt = true
	}

	// <-??
}
