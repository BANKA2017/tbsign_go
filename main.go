package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_api "github.com/BANKA2017/tbsign_go/api"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"gorm.io/gorm/logger"
)

// install
var setup bool
var autoInstall bool

var _adminName string
var _adminEmail string
var _adminPassword string

var EncryptDataAction string

var err error

func init() {
	fmt.Println("TbSign➡️\n--- info ---")
	fmt.Println("build_at:", share.BuiltAt)
	fmt.Println("commit_hash:", share.BuildGitCommitHash)
	fmt.Println("frontend_hash:", share.BuildEmbeddedFrontendGitCommitHash)
	fmt.Println("publish_type:", share.BuildPublishType)
	if share.BuiltAt != "Now" && share.BuildGitCommitHash != "N/A" && share.BuildEmbeddedFrontendGitCommitHash != "N/A" {
		fmt.Println("version:", fmt.Sprintf("%s.%s.%s\n", share.BuildAtTime.Format("20060102"), share.BuildGitCommitHash[0:7], share.BuildEmbeddedFrontendGitCommitHash[0:7]))
	} else {
		fmt.Println("version: dev\n------------")
	}
}

func main() {
	// sqlite
	flag.StringVar(&share.DBPath, "db_path", "", "Database path")

	// mysql
	var tmpHost string
	flag.StringVar(&share.DBUsername, "username", "", "Username")
	flag.StringVar(&share.DBPassword, "pwd", "", "Password")
	flag.StringVar(&tmpHost, "endpoint", "127.0.0.1:3306", "MySQL host:port (deprecated)")
	flag.StringVar(&share.DBEndpoint, "host", "127.0.0.1:3306", "MySQL host:port")
	flag.StringVar(&share.DBName, "db", "tbsign", "Database name")
	flag.StringVar(&share.DBTLSOption, "db_tls", "false", "Option for CA cert (MySQL only)")
	flag.StringVar(&share.DBMode, "db_mode", "mysql", "sqlite/mysql/pgsql")

	//proxy
	flag.BoolVar(&_function.IgnoreProxy, "no_proxy", false, "Ignore the http proxy config from environment vars")

	// api
	flag.BoolVar(&share.EnableApi, "api", false, "active backend endpoints")
	flag.BoolVar(&share.EnableFrontend, "fe", false, "active frontend endpoints")
	flag.BoolVar(&share.EnableBackup, "allow_backup", false, "allow backup (export/import)")
	flag.StringVar(&share.Address, "address", ":1323", "address :1323")

	// setup
	flag.BoolVar(&setup, "setup", false, "Init the system [force]")
	flag.BoolVar(&autoInstall, "auto_install", false, "Auto install the system when tables are not exist")
	flag.StringVar(&_adminName, "admin_name", "", "Name of admin")
	flag.StringVar(&_adminEmail, "admin_email", "", "Email of admin")
	flag.StringVar(&_adminPassword, "admin_password", "", "Password of admin")

	// --experimental-*
	// encrypt
	flag.StringVar(&EncryptDataAction, "data_encrypt_action", "", "Encrypt/Decrypt data in database")
	flag.StringVar(&share.DataEncryptKeyStr, "data_encrypt_key", "", "The key to encrypt some user data (base64url)")
	// flag.BoolVar(&share.DisableEmail, "disable-email", false, "disable email")
	flag.StringVar(&share.DNSAddress, "dns_addr", "", "DNS Address")

	// others
	flag.BoolVar(&share.TestMode, "test", false, "Not send any requests to tieba servers")

	flag.Parse()

	if share.DBEndpoint == "127.0.0.1:3306" && tmpHost != "127.0.0.1:3306" {
		share.DBEndpoint = tmpHost
	}

	// from env
	if share.DBUsername == "" {
		share.DBUsername = os.Getenv("tc_username")
	}
	if share.DBPassword == "" {
		share.DBPassword = os.Getenv("tc_pwd")
	}
	if share.DBEndpoint == "127.0.0.1:3306" && os.Getenv("tc_host") != "" {
		share.DBEndpoint = os.Getenv("tc_host")
	}
	if share.DBEndpoint == "127.0.0.1:3306" && os.Getenv("tc_endpoint") != "" {
		share.DBEndpoint = os.Getenv("tc_endpoint")
	}
	if share.DBName == "tbsign" && os.Getenv("tc_db") != "" {
		share.DBName = os.Getenv("tc_db")
	}

	// DBMode **must** be set before DBTLSOption
	if share.DBMode == "mysql" && os.Getenv("tc_db_mode") != "" {
		share.DBMode = os.Getenv("tc_db_mode")
	}
	if share.DBTLSOption == "false" && os.Getenv("tc_db_tls") != "" {
		share.DBTLSOption = os.Getenv("tc_db_tls")
	}

	if share.DBTLSOption == "false" && share.DBMode == "pgsql" {
		share.DBTLSOption = "prefer" // default value
	}

	if share.DBPath == "" && os.Getenv("tc_db_path") != "" {
		share.DBPath = os.Getenv("tc_db_path")
	}
	if !share.TestMode && os.Getenv("tc_test") != "" {
		share.TestMode = os.Getenv("tc_test") == "true"
	}
	if !share.EnableApi && os.Getenv("tc_api") != "" {
		share.EnableApi = os.Getenv("tc_api") == "true"
	}
	if !share.EnableFrontend && os.Getenv("tc_fe") != "" {
		share.EnableFrontend = os.Getenv("tc_fe") == "true"
	}

	if !share.EnableApi && share.EnableFrontend {
		log.Fatal("ERROR: 不允许关闭 api 的同时又启用前端!!!")
	}

	if !share.EnableBackup && os.Getenv("tc_allow_backup") != "" {
		share.EnableBackup = os.Getenv("tc_allow_backup") == "true"
	}

	if share.DataEncryptKeyStr == "" && os.Getenv("tc_data_encrypt_key") != "" {
		share.DataEncryptKeyStr = os.Getenv("tc_data_encrypt_key")
	}
	if share.DataEncryptKeyStr != "" {
		share.DataEncryptKeyByte, err = _function.Base64URLDecode(share.DataEncryptKeyStr)
		if err != nil {
			log.Fatal(err)
		}
		if len(share.DataEncryptKeyByte) != 32 {
			log.Fatal("ERROR: 密钥长度无效")
		}
	}

	if share.DNSAddress == "" && os.Getenv("tc_dns_addr") != "" {
		share.DNSAddress = os.Getenv("tc_dns_addr")
	}

	if share.Address == ":1323" && os.Getenv("tc_address") != "" {
		share.Address = os.Getenv("tc_address")
	}

	if !autoInstall && os.Getenv("tc_auto_install") != "" {
		autoInstall = true
	}
	if _adminName == "" && os.Getenv("tc_admin_name") != "" {
		_adminName = os.Getenv("tc_admin_name")
	}
	if _adminEmail == "" && os.Getenv("tc_admin_email") != "" {
		_adminEmail = os.Getenv("tc_admin_email")
	}
	if _adminPassword == "" && os.Getenv("tc_admin_password") != "" {
		_adminPassword = os.Getenv("tc_admin_password")
	}

	if setup {
		log.Println("WARNING: 覆盖安装已启用，会覆盖现有数据，请做好备份")
	}

	if setup && autoInstall {
		log.Fatal("ERROR: 不允许自动化覆盖安装!!!")
	} else if autoInstall && _adminName != "" && _adminEmail != "" && _adminPassword != "" {
		log.Println("WARNING: 已启用自动安装")
	} else if autoInstall {
		log.Fatal("ERROR: 管理员信息不完整，无法安装")
	}

	// connect to db
	logLevel := logger.Error
	if share.TestMode {
		logLevel = logger.Info
	}

	dbExists := true
	versionStruct := struct {
		Version string
	}{}

	_function.GormDB.LogLevel = logLevel
	_function.GormDB.ServicePrefix = "tbsign-db"
	_function.GormDB.WALMode = true

	if share.DBMode == "pgsql" {
		// pgsql
		if share.DBUsername == "" {
			log.Fatal("global: Empty username")
		}
		// precheck table
		if err = _function.GormDB.ConnectToPostgreSQL(share.DBUsername, share.DBPassword, share.DBEndpoint, "", share.DBTLSOption); err != nil {
			log.Fatal("db:", err)
		}

		var count struct {
			Count int64
		}

		_function.GormDB.R.Raw("SELECT COUNT(*) AS count FROM pg_database WHERE datname = ?;", share.DBName).Scan(&count)
		dbExists = count.Count > 0
		if !dbExists {
			log.Println("db:", share.DBName, "is not exists")
			setup = true
		}

		// setup
		if setup {
			_plugin.SetupSystem(share.DBMode, "", share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, share.DBTLSOption, logLevel, dbExists, autoInstall, _adminName, _adminEmail, _adminPassword)
		} else {
			sqlDB, _ := _function.GormDB.R.DB()
			sqlDB.Close()

			if err = _function.GormDB.ConnectToPostgreSQL(share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, share.DBTLSOption); err != nil {
				log.Fatal("db:", err)
			}
		}

		// version

		_function.GormDB.R.Raw("SELECT version();").Scan(&versionStruct)
		share.DBVersion = versionStruct.Version
	} else if share.DBPath != "" {
		// sqlite
		share.DBMode = "sqlite"
		if _, err := os.Stat(share.DBPath); err != nil && os.IsNotExist(err) {
			dbExists = false
			setup = true
		}

		if err = _function.GormDB.ConnectToSQLite(share.DBPath); err != nil {
			log.Fatal("db:", err)
		}

		// setup
		if setup {
			_plugin.SetupSystem(share.DBMode, share.DBPath, "", "", "", "", share.DBTLSOption, logLevel, dbExists, autoInstall, _adminName, _adminEmail, _adminPassword)
		}

		_function.GormDB.R.Raw("SELECT sqlite_version() AS version;").Scan(&versionStruct)
		share.DBVersion = versionStruct.Version
	} else {
		share.DBMode = "mysql"
		// mysql
		if share.DBUsername == "" {
			log.Fatal("global: Empty username")
		}
		// precheck table

		if err = _function.GormDB.ConnectToMySQL(share.DBUsername, share.DBPassword, share.DBEndpoint, "", share.DBTLSOption); err != nil {
			log.Fatal("db:", err)
		}

		var count struct {
			Count int64
		}

		_function.GormDB.R.Raw("SELECT (COUNT(*) > 0) AS count FROM information_schema.tables WHERE table_schema = ?;", share.DBName).Scan(&count)
		dbExists = count.Count > 0
		if !dbExists {
			log.Println("db:", share.DBName, "is not exists")
			setup = true
		}

		// setup
		if setup {
			_plugin.SetupSystem(share.DBMode, "", share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, share.DBTLSOption, logLevel, dbExists, autoInstall, _adminName, _adminEmail, _adminPassword)
		} else {
			sqlDB, _ := _function.GormDB.R.DB()
			sqlDB.Close()

			if err = _function.GormDB.ConnectToMySQL(share.DBUsername, share.DBPassword, share.DBEndpoint, share.DBName, share.DBTLSOption); err != nil {
				log.Fatal("db:", err)
			}
		}

		// version

		_function.GormDB.R.Raw("SELECT @@version AS version;").Scan(&versionStruct)
		share.DBVersion = versionStruct.Version
	}

	// init
	_function.InitOptions()
	share.IsPureGO = _function.GetOption("go_ver") == "1"
	share.IsEncrypt = _function.GetOption("go_encrypt") != "0"
	_plugin.InitPluginList()

	// encrypt/decrypt init
	if share.IsPureGO && EncryptDataAction != "" {
		if len(share.DataEncryptKeyByte) != 32 {
			log.Fatal("ERROR: 无效密钥，无法加密/解密")
		} else if strings.EqualFold(EncryptDataAction, "encrypt") && len(share.DataEncryptKeyByte) > 0 {
			err := _plugin.EncryptBaiduIDData()
			if err != nil {
				log.Fatal(err)
			}
			log.Println("INFO: 加密完成")
			os.Exit(0)
		} else if strings.EqualFold(EncryptDataAction, "decrypt") && len(share.DataEncryptKeyByte) > 0 {
			err := _plugin.DecryptBaiduIDData()
			if err != nil {
				log.Fatal(err)
			}
			log.Println("INFO: 解密完成")
			os.Exit(0)
		}
	} else if len(share.DataEncryptKeyByte) == 32 && !share.IsPureGO {
		// DO NOT USE ENCRYPT IN COMPAT MODE!!!
		share.DataEncryptKeyByte = []byte{}
		share.DataEncryptKeyStr = ""

		log.Println("WARNING: 兼容模式下不支持加密，已恢复使用明文")
	}

	if share.IsEncrypt && len(share.DataEncryptKeyByte) != 32 {
		log.Fatal("ERROR: 无效密钥，无法加密/解密数据")
	}
	if !share.IsEncrypt && len(share.DataEncryptKeyByte) > 0 {
		share.DataEncryptKeyByte = []byte{}
		share.DataEncryptKeyStr = ""

		log.Println("WARNING: 数据未加密，已恢复使用明文")
	}

	// log.Println(share.DBVersion)

	/// client
	/// DO NOT EXEC _function.InitClient BEFORE READING FLAGS AND ENV!!!!!
	_function.DefaultCient = _function.InitClient(300 * time.Second)
	_function.TBClient = _function.InitClient(10 * time.Second)

	if share.EnableApi {
		go _api.Api(share.Address)
	}

	// Interval
	oneSecondInterval := time.NewTicker(time.Second)
	defer oneSecondInterval.Stop()
	oneMinuteInterval := time.NewTicker(time.Minute)
	defer oneMinuteInterval.Stop()
	fourHoursInterval := time.NewTicker(time.Hour * 4)
	defer fourHoursInterval.Stop()

	// cron
	for {
		select {
		case <-oneSecondInterval.C:
			_function.UpdateNow()
		case <-oneMinuteInterval.C:
			if share.TestMode {
				continue
			}
			_plugin.DoCheckinAction()
			_plugin.DoReCheckinAction()

			// plugins
			for _, info := range _plugin.PluginList {
				if _function.TinyIntToBool(info.(_plugin.PluginHooks).GetDBInfo().Status) {
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
